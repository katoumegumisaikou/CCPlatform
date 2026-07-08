package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	platformmqtt "ccplatform/internal/mqtt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	workIdle     int8 = 0
	workCleaning int8 = 1
	workCharging int8 = 2
	workFault    int8 = 3
)

type config struct {
	broker        string
	robots        stringList
	prefix        string
	count         int
	mode          string
	interval      time.Duration
	statusEvery   time.Duration
	sensorEvery   time.Duration
	alarmEvery    time.Duration
	duration      time.Duration
	qos           uint
	startX        float64
	startY        float64
	gpsLongitude  float64
	gpsLatitude   float64
	taskID        uint
	clientID      string
	cleanAreaStep float64
}

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*s = append(*s, part)
		}
	}
	return nil
}

type robotSim struct {
	mu        sync.Mutex
	id        string
	rng       *rand.Rand
	mode      string
	work      int8
	taskID    uint
	posX      float64
	posY      float64
	posZ      float64
	heading   float64
	speed     float64
	cleanArea float64
	battery   int8
	faultCode int
	startedAt time.Time
}

type commandPayload struct {
	Cmd    string                 `json:"cmd"`
	Params map[string]interface{} `json:"params"`
}

func main() {
	cfg := parseFlags()
	robotIDs := cfg.robotIDs()
	if len(robotIDs) == 0 {
		log.Fatal("no robots configured")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.duration)
		defer cancel()
	}

	sims := make(map[string]*robotSim, len(robotIDs))
	for i, id := range robotIDs {
		sims[id] = newRobotSim(id, i, cfg)
	}

	client := connectMQTT(cfg, sims)
	defer client.Disconnect(250)

	log.Printf("simulator started broker=%s robots=%s mode=%s", cfg.broker, strings.Join(robotIDs, ","), cfg.mode)
	run(ctx, client, cfg, sims)
	log.Print("simulator stopped")
}

func parseFlags() config {
	cfg := config{
		broker:        "tcp://127.0.0.1:1883",
		prefix:        "SIM-R",
		mode:          "clean",
		interval:      time.Second,
		statusEvery:   10 * time.Second,
		sensorEvery:   30 * time.Second,
		alarmEvery:    0,
		startX:        20,
		startY:        20,
		gpsLongitude:  113.3245,
		gpsLatitude:   23.1065,
		clientID:      fmt.Sprintf("ccplatform-simrobot-%d", time.Now().UnixNano()),
		cleanAreaStep: 0.8,
	}

	flag.StringVar(&cfg.broker, "broker", cfg.broker, "MQTT broker URL")
	flag.Var(&cfg.robots, "robots", "comma-separated robot IDs; repeatable")
	flag.StringVar(&cfg.prefix, "prefix", cfg.prefix, "robot ID prefix used with -count when -robots is empty")
	flag.IntVar(&cfg.count, "count", 0, "robot count generated from -prefix when -robots is empty")
	flag.StringVar(&cfg.mode, "mode", cfg.mode, "simulation mode: idle, clean, fault, mixed")
	flag.DurationVar(&cfg.interval, "interval", cfg.interval, "heartbeat and position interval")
	flag.DurationVar(&cfg.statusEvery, "status-every", cfg.statusEvery, "status publish interval")
	flag.DurationVar(&cfg.sensorEvery, "sensor-every", cfg.sensorEvery, "sensor publish interval; 0 disables sensors")
	flag.DurationVar(&cfg.alarmEvery, "alarm-every", cfg.alarmEvery, "alarm publish interval; 0 disables scheduled alarms")
	flag.DurationVar(&cfg.duration, "duration", cfg.duration, "run duration; 0 runs until interrupted")
	flag.UintVar(&cfg.qos, "qos", 0, "publish QoS")
	flag.Float64Var(&cfg.startX, "start-x", cfg.startX, "initial X coordinate")
	flag.Float64Var(&cfg.startY, "start-y", cfg.startY, "initial Y coordinate")
	flag.Float64Var(&cfg.gpsLongitude, "gps-lng", cfg.gpsLongitude, "base GPS longitude")
	flag.Float64Var(&cfg.gpsLatitude, "gps-lat", cfg.gpsLatitude, "base GPS latitude")
	flag.UintVar(&cfg.taskID, "task-id", 0, "cleaning task ID included in clean messages")
	flag.StringVar(&cfg.clientID, "client-id", cfg.clientID, "MQTT client ID")
	flag.Float64Var(&cfg.cleanAreaStep, "clean-step", cfg.cleanAreaStep, "clean area increment per cleaning tick")
	flag.Parse()

	cfg.mode = strings.ToLower(strings.TrimSpace(cfg.mode))
	switch cfg.mode {
	case "idle", "clean", "fault", "mixed":
	default:
		log.Fatalf("invalid -mode %q", cfg.mode)
	}
	if cfg.interval <= 0 {
		log.Fatal("-interval must be positive")
	}
	if cfg.statusEvery <= 0 {
		log.Fatal("-status-every must be positive")
	}
	if cfg.qos > 2 {
		log.Fatal("-qos must be 0, 1, or 2")
	}
	return cfg
}

func (cfg config) robotIDs() []string {
	seen := map[string]bool{}
	var ids []string
	for _, id := range cfg.robots {
		id = strings.TrimSpace(id)
		if id != "" && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	if len(ids) > 0 {
		return ids
	}
	for i := 1; i <= cfg.count; i++ {
		ids = append(ids, fmt.Sprintf("%s%04d", cfg.prefix, i))
	}
	if len(ids) == 0 {
		ids = append(ids, "R1-0001")
	}
	return ids
}

func newRobotSim(id string, index int, cfg config) *robotSim {
	work := workCleaning
	speed := 0.8
	faultCode := 0
	switch cfg.mode {
	case "idle":
		work = workIdle
		speed = 0
	case "fault":
		work = workFault
		speed = 0
		faultCode = 1001
	case "mixed":
		switch index % 4 {
		case 0:
			work = workCleaning
			speed = 0.8
		case 1:
			work = workIdle
			speed = 0
		case 2:
			work = workCharging
			speed = 0
		default:
			work = workFault
			speed = 0
			faultCode = 1001
		}
	}

	return &robotSim{
		id:        id,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano() + int64(index)*9973)),
		mode:      cfg.mode,
		work:      work,
		taskID:    cfg.taskID,
		posX:      cfg.startX + float64(index)*6,
		posY:      cfg.startY + float64(index%3)*4,
		heading:   float64((index * 35) % 360),
		speed:     speed,
		battery:   int8(86 - index%20),
		faultCode: faultCode,
		startedAt: time.Now(),
	}
}

func connectMQTT(cfg config, sims map[string]*robotSim) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.broker).
		SetClientID(cfg.clientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(time.Second)

	opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
		handleDownstream(sims, msg.Topic(), msg.Payload())
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); !token.WaitTimeout(10*time.Second) || token.Error() != nil {
		if token.Error() != nil {
			log.Fatalf("connect MQTT broker: %v", token.Error())
		}
		log.Fatal("connect MQTT broker: timeout")
	}

	for id := range sims {
		for _, topic := range []string{fmt.Sprintf("tdw/robot/%s/cmd", id), fmt.Sprintf("tdw/robot/%s/config", id)} {
			if token := client.Subscribe(topic, byte(cfg.qos), nil); !token.WaitTimeout(5*time.Second) || token.Error() != nil {
				if token.Error() != nil {
					log.Fatalf("subscribe %s: %v", topic, token.Error())
				}
				log.Fatalf("subscribe %s: timeout", topic)
			}
		}
	}
	return client
}

func handleDownstream(sims map[string]*robotSim, topic string, payload []byte) {
	parts := strings.Split(topic, "/")
	if len(parts) != 4 || parts[0] != "tdw" || parts[1] != "robot" {
		return
	}
	sim := sims[parts[2]]
	if sim == nil {
		return
	}

	switch parts[3] {
	case "cmd":
		var cmd commandPayload
		if err := json.Unmarshal(payload, &cmd); err != nil {
			log.Printf("[%s] invalid command payload: %v", sim.id, err)
			return
		}
		sim.applyCommand(cmd)
		log.Printf("[%s] command=%s params=%s", sim.id, cmd.Cmd, compactJSON(cmd.Params))
	case "config":
		log.Printf("[%s] config=%s", sim.id, string(payload))
	}
}

func (r *robotSim) applyCommand(cmd commandPayload) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch strings.ToLower(cmd.Cmd) {
	case "start":
		r.work = workCleaning
		r.speed = 0.8
		r.faultCode = 0
	case "stop":
		r.work = workIdle
		r.speed = 0
	case "return", "charge":
		r.work = workCharging
		r.speed = 0
	case "reset":
		r.work = workIdle
		r.speed = 0
		r.faultCode = 0
	case "fault":
		r.work = workFault
		r.speed = 0
		r.faultCode = 1001
	case "task":
		if id, ok := uintParam(cmd.Params, "task_id"); ok {
			r.taskID = id
		}
		r.work = workCleaning
		r.speed = 0.8
		r.faultCode = 0
	}
}

func run(ctx context.Context, client mqtt.Client, cfg config, sims map[string]*robotSim) {
	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	statusTicker := time.NewTicker(cfg.statusEvery)
	defer statusTicker.Stop()

	var sensorTicker *time.Ticker
	var sensorC <-chan time.Time
	if cfg.sensorEvery > 0 {
		sensorTicker = time.NewTicker(cfg.sensorEvery)
		defer sensorTicker.Stop()
		sensorC = sensorTicker.C
	}

	var alarmTicker *time.Ticker
	var alarmC <-chan time.Time
	if cfg.alarmEvery > 0 {
		alarmTicker = time.NewTicker(cfg.alarmEvery)
		defer alarmTicker.Stop()
		alarmC = alarmTicker.C
	}

	publishAll(client, cfg, sims, true, true, cfg.sensorEvery > 0, shouldPublishAlarm(sims))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			publishAll(client, cfg, sims, true, false, false, false)
		case <-statusTicker.C:
			publishAll(client, cfg, sims, false, true, false, false)
		case <-sensorC:
			publishAll(client, cfg, sims, false, false, true, false)
		case <-alarmC:
			publishAll(client, cfg, sims, false, false, false, true)
		}
	}
}

func shouldPublishAlarm(sims map[string]*robotSim) bool {
	for _, sim := range sims {
		sim.mu.Lock()
		isFault := sim.work == workFault
		sim.mu.Unlock()
		if isFault {
			return true
		}
	}
	return false
}

func publishAll(client mqtt.Client, cfg config, sims map[string]*robotSim, movement, status, sensor, alarm bool) {
	for _, sim := range sims {
		sim.advance(cfg)
		if movement {
			publish(client, byte(cfg.qos), topic(sim.id, "heartbeat"), sim.heartbeat())
			publish(client, byte(cfg.qos), topic(sim.id, "position"), sim.position(cfg))
			if sim.isCleaning() {
				publish(client, byte(cfg.qos), topic(sim.id, "clean"), sim.clean())
			}
		}
		if status {
			publish(client, byte(cfg.qos), topic(sim.id, "status"), sim.status())
		}
		if sensor {
			publish(client, byte(cfg.qos), topic(sim.id, "sensor"), sim.sensor())
		}
		if alarm && sim.isFault() {
			publish(client, byte(cfg.qos), topic(sim.id, "alarm"), sim.alarm())
		}
	}
}

func (r *robotSim) advance(cfg config) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.work == workCleaning {
		distance := r.speed * cfg.interval.Seconds()
		r.heading = math.Mod(r.heading+2+r.rng.Float64()*4, 360)
		r.posX += math.Cos(r.heading*math.Pi/180) * distance
		r.posY += math.Sin(r.heading*math.Pi/180) * distance
		r.cleanArea += cfg.cleanAreaStep + r.rng.Float64()*0.4
		if r.battery > 10 && r.rng.Intn(20) == 0 {
			r.battery--
		}
		return
	}
	if r.work == workCharging && r.battery < 100 && r.rng.Intn(4) == 0 {
		r.battery++
	}
}

func (r *robotSim) heartbeat() platformmqtt.HeartbeatPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	return platformmqtt.HeartbeatPayload{
		RobotID:      r.id,
		Timestamp:    time.Now().Unix(),
		BatteryLevel: r.battery,
		OnlineStatus: 1,
		WorkStatus:   r.work,
	}
}

func (r *robotSim) position(cfg config) platformmqtt.PositionPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	lng := cfg.gpsLongitude + r.posX/100000
	lat := cfg.gpsLatitude + r.posY/100000
	alt := 12 + r.posZ
	acc := 1.5 + r.rng.Float64()
	return platformmqtt.PositionPayload{
		RobotID:      r.id,
		Timestamp:    time.Now().Unix(),
		PosX:         round(r.posX),
		PosY:         round(r.posY),
		PosZ:         round(r.posZ),
		Heading:      round(r.heading),
		GPSLongitude: &lng,
		GPSLatitude:  &lat,
		GPSAltitude:  &alt,
		GPSAccuracy:  &acc,
	}
}

func (r *robotSim) status() platformmqtt.StatusPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	runDuration := int64(now.Sub(r.startedAt).Seconds())
	batteryVoltage := 48.0 + float64(r.battery)/100*6
	faultStatus := int8(0)
	if r.work == workFault {
		faultStatus = 1
	}
	workPeriod := int8(1)
	signal := 75 + r.rng.Intn(20)
	signal4G := 70 + r.rng.Intn(25)
	cartBattery := batteryVoltage - 1.2
	shuttleBattery := batteryVoltage + 0.8
	cartLow := int8(0)
	shuttleLow := int8(0)
	shuttleCurrent := 8.0 + r.rng.Float64()*2
	cleanCurrent := 5.0 + r.rng.Float64()*1.5
	travelCurrent := 6.0 + r.rng.Float64()*2
	shuttlePower := int8(60)
	startPosition := int8(1)
	endPosition := int8(2)
	hasCleaner := int8(1)
	deviceTime := now.Unix()

	return platformmqtt.StatusPayload{
		RobotID:                 r.id,
		Timestamp:               now.Unix(),
		WorkStatus:              r.work,
		Speed:                   round(r.speed),
		CleanArea:               round(r.cleanArea),
		FaultCode:               r.faultCode,
		Temperature:             round(24 + r.rng.Float64()*8),
		Humidity:                round(45 + r.rng.Float64()*20),
		LightIntensity:          round(45000 + r.rng.Float64()*20000),
		WindSpeed:               round(2 + r.rng.Float64()*4),
		BatteryVoltage:          &batteryVoltage,
		FaultStatus:             &faultStatus,
		WorkPeriod:              &workPeriod,
		SignalStrength:          &signal,
		Signal4GStrength:        &signal4G,
		RunDurationSeconds:      &runDuration,
		CartBatteryVoltage:      &cartBattery,
		CartLowVoltageStatus:    &cartLow,
		ShuttleBatteryVoltage:   &shuttleBattery,
		ShuttleLowVoltageStatus: &shuttleLow,
		ShuttleMotorCurrent:     &shuttleCurrent,
		CartCleanMotorCurrent:   &cleanCurrent,
		CartTravelMotorCurrent:  &travelCurrent,
		ShuttlePowerPercent:     &shuttlePower,
		ShuttleStartPosition:    &startPosition,
		ShuttleEndPosition:      &endPosition,
		ShuttleHasCleaner:       &hasCleaner,
		DeviceTime:              &deviceTime,
	}
}

func (r *robotSim) clean() platformmqtt.CleanPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	return platformmqtt.CleanPayload{
		RobotID:   r.id,
		TaskID:    r.taskID,
		Timestamp: time.Now().Unix(),
		PosX:      round(r.posX),
		PosY:      round(r.posY),
		CleanArea: round(r.cleanArea),
	}
}

func (r *robotSim) sensor() platformmqtt.SensorPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	rms := 0.4 + r.rng.Float64()*0.2
	peak := rms * (2.8 + r.rng.Float64())
	windingTemp := 46 + r.rng.Float64()*8
	controllerTemp := 42 + r.rng.Float64()*6
	insulation := 80 + r.rng.Float64()*30
	component := "drive_motor"
	if r.rng.Intn(2) == 0 {
		component = "brush_motor"
	}
	samples := make([]float64, 32)
	for i := range samples {
		samples[i] = round(math.Sin(float64(i)/3)*rms + (r.rng.Float64()-0.5)*0.08)
	}
	return platformmqtt.SensorPayload{
		RobotID:              r.id,
		Timestamp:            time.Now().Unix(),
		Component:            component,
		VibrationSamples:     samples,
		SampleRateHz:         100,
		VibrationRMS:         &rms,
		VibrationPeak:        &peak,
		WindingTemp:          &windingTemp,
		ControllerTemp:       &controllerTemp,
		InsulationResistMOhm: &insulation,
	}
}

func (r *robotSim) alarm() platformmqtt.AlarmPayload {
	return platformmqtt.AlarmPayload{
		RobotID:      r.id,
		Timestamp:    time.Now().Unix(),
		AlarmLevel:   3,
		AlarmType:    "SIM_FAULT",
		AlarmContent: "simulated robot fault",
	}
}

func (r *robotSim) isCleaning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.work == workCleaning
}

func (r *robotSim) isFault() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.work == workFault
}

func publish(client mqtt.Client, qos byte, topic string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal %s: %v", topic, err)
		return
	}
	token := client.Publish(topic, qos, false, data)
	if !token.WaitTimeout(5 * time.Second) {
		log.Printf("publish %s timeout", topic)
		return
	}
	if err := token.Error(); err != nil {
		log.Printf("publish %s: %v", topic, err)
	}
}

func topic(robotID, msgType string) string {
	return fmt.Sprintf("tdw/robot/%s/%s", robotID, msgType)
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

func uintParam(params map[string]interface{}, key string) (uint, bool) {
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return uint(v), v >= 0
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		return uint(n), err == nil
	default:
		return 0, false
	}
}

func compactJSON(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
