# MQTT robot simulator

`simrobot` publishes robot telemetry to the existing CCPlatform MQTT topics:

- `tdw/robot/{robotID}/heartbeat`
- `tdw/robot/{robotID}/position`
- `tdw/robot/{robotID}/status`
- `tdw/robot/{robotID}/clean`
- `tdw/robot/{robotID}/sensor`
- `tdw/robot/{robotID}/alarm`

It also subscribes to the downstream topics used by the platform:

- `tdw/robot/{robotID}/cmd`
- `tdw/robot/{robotID}/config`

## Usage

Run one existing robot against the local broker:

```bash
go run ./cmd/simrobot -robots R1-0001 -mode clean
```

Run multiple existing robots:

```bash
go run ./cmd/simrobot -robots R1-0001,R1-0002,R1-0003 -mode mixed
```

Run generated robot IDs:

```bash
go run ./cmd/simrobot -prefix SIM-R -count 5 -mode clean
```

Generated IDs are useful for MQTT traffic tests. For database, alarm, cleaning
record, and frontend monitor flows, prefer robot IDs that already exist in the
local database because the platform updates existing robot rows.

Useful flags:

- `-broker tcp://127.0.0.1:1883`
- `-duration 30s`
- `-interval 1s`
- `-status-every 10s`
- `-sensor-every 30s`
- `-alarm-every 1m`
- `-mode idle|clean|fault|mixed`
- `-task-id 123`
