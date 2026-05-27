import { useEffect, useRef, type CSSProperties } from 'react';

interface SceneRobot {
  robot_id: string;
  robot_name: string;
  robot_type: number;
  online_status: number;
  work_status: number;
  battery_level: number;
  pos_x: number;
  pos_y: number;
  heading: number;
  speed: number;
  clean_area: number;
}

interface Robot2DSceneProps {
  robot: SceneRobot;
  peers?: SceneRobot[];
  trailPoints?: Array<{
    pos_x: number;
    pos_y: number;
  }>;
  footerLabel?: string;
  style?: CSSProperties;
}

const BACKGROUND_TOP = '#eaf6ff';
const BACKGROUND_BOTTOM = '#f7f3df';
const PANEL_FILL = '#244f7a';
const PANEL_EDGE = '#8bd3ff';
const ONLINE_GLOW = '#37c776';
const OFFLINE_GLOW = '#bfbfbf';
const STATUS_COLORS: Record<number, string> = {
  0: '#8c8c8c',
  1: '#00a6ff',
  2: '#ffb648',
  3: '#ff5a5f',
  4: '#8758ff',
};

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function lerp(from: number, to: number, amount: number) {
  return from + (to - from) * amount;
}

function lerpAngle(from: number, to: number, amount: number) {
  let delta = ((to - from + 540) % 360) - 180;
  if (delta < -180) delta += 360;
  return from + delta * amount;
}

function normalize(value: number, min: number, max: number, padding: number) {
  if (!Number.isFinite(value)) return 0.5;
  if (Math.abs(max - min) < 0.0001) return 0.5;
  return padding + ((value - min) / (max - min)) * (1 - padding * 2);
}

function drawSolarField(ctx: CanvasRenderingContext2D, width: number, height: number, phase: number) {
  const sky = ctx.createLinearGradient(0, 0, 0, height);
  sky.addColorStop(0, BACKGROUND_TOP);
  sky.addColorStop(1, BACKGROUND_BOTTOM);
  ctx.fillStyle = sky;
  ctx.fillRect(0, 0, width, height);

  ctx.fillStyle = 'rgba(255,255,255,0.55)';
  ctx.beginPath();
  ctx.arc(width - 72, 58, 26, 0, Math.PI * 2);
  ctx.fill();

  for (let i = 0; i < 6; i += 1) {
    const laneTop = 52 + i * 34;
    ctx.strokeStyle = i % 2 === 0 ? 'rgba(36,79,122,0.1)' : 'rgba(36,79,122,0.06)';
    ctx.lineWidth = 20;
    ctx.beginPath();
    ctx.moveTo(18, laneTop);
    ctx.lineTo(width - 18, laneTop + 8);
    ctx.stroke();

    ctx.fillStyle = PANEL_FILL;
    ctx.strokeStyle = PANEL_EDGE;
    ctx.lineWidth = 1;
    for (let col = 0; col < 7; col += 1) {
      const x = 22 + col * ((width - 44) / 7);
      const y = laneTop - 10 + (col % 2) * 2;
      const w = (width - 78) / 7;
      const h = 18;
      ctx.fillRect(x, y, w, h);
      ctx.strokeRect(x, y, w, h);
    }
  }

  ctx.strokeStyle = 'rgba(255,255,255,0.22)';
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(0, height * 0.75);
  for (let x = 0; x <= width; x += 18) {
    const y = height * 0.75 + Math.sin((x + phase * 90) / 28) * 4;
    ctx.lineTo(x, y);
  }
  ctx.stroke();
}

function drawRobotBody(
  ctx: CanvasRenderingContext2D,
  robotType: number,
  color: string,
  energy: number,
) {
  ctx.save();
  ctx.fillStyle = color;
  ctx.strokeStyle = 'rgba(255,255,255,0.92)';
  ctx.lineWidth = 2;

  if (robotType === 1) {
    ctx.beginPath();
    ctx.roundRect(-24, -12, 48, 24, 8);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = 'rgba(255,255,255,0.24)';
    ctx.fillRect(-14, -18, 28, 7);
    ctx.fillStyle = '#cfd8dc';
    ctx.fillRect(-30, 10, 60, 4);
  } else if (robotType === 2) {
    ctx.beginPath();
    ctx.moveTo(-28, 0);
    ctx.lineTo(-10, -18);
    ctx.lineTo(16, -18);
    ctx.lineTo(28, 0);
    ctx.lineTo(12, 18);
    ctx.lineTo(-18, 18);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = '#dde7ef';
    ctx.fillRect(-36, 16, 72, 5);
  } else {
    ctx.beginPath();
    ctx.arc(0, 0, 22, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = 'rgba(255,255,255,0.2)';
    ctx.beginPath();
    ctx.arc(0, 0, 10 + energy * 4, 0, Math.PI * 2);
    ctx.fill();
  }

  ctx.fillStyle = '#ffffff';
  ctx.beginPath();
  ctx.moveTo(8, 0);
  ctx.lineTo(-6, -8);
  ctx.lineTo(-2, 0);
  ctx.lineTo(-6, 8);
  ctx.closePath();
  ctx.fill();
  ctx.restore();
}

export default function Robot2DScene({
  robot,
  peers = [],
  trailPoints = [],
  footerLabel,
  style,
}: Robot2DSceneProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const latestRobotRef = useRef(robot);
  const latestPeersRef = useRef(peers);
  const latestTrailRef = useRef(trailPoints);
  const latestFooterLabelRef = useRef(footerLabel);
  const animatedRef = useRef({ x: 0.5, y: 0.5, heading: 0 });
  const trailRef = useRef<Array<{ x: number; y: number }>>([]);

  useEffect(() => {
    latestRobotRef.current = robot;
  }, [robot]);

  useEffect(() => {
    latestPeersRef.current = peers;
  }, [peers]);

  useEffect(() => {
    latestTrailRef.current = trailPoints;
  }, [trailPoints]);

  useEffect(() => {
    latestFooterLabelRef.current = footerLabel;
  }, [footerLabel]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return undefined;

    const context = canvas.getContext('2d');
    if (!context) return undefined;

    let frameId = 0;

    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;
      canvas.width = Math.round(rect.width * dpr);
      canvas.height = Math.round(rect.height * dpr);
      context.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    resize();
    const observer = new ResizeObserver(resize);
    observer.observe(canvas);

    const render = (time: number) => {
      const width = canvas.clientWidth;
      const height = canvas.clientHeight;
      const phase = time / 1000;
      const currentRobot = latestRobotRef.current;
      const currentPeers = latestPeersRef.current.length > 0 ? latestPeersRef.current : [currentRobot];
      const externalTrail = latestTrailRef.current;

      const xValues = currentPeers.map((item) => item.pos_x);
      const yValues = currentPeers.map((item) => item.pos_y);
      xValues.push(currentRobot.pos_x);
      yValues.push(currentRobot.pos_y);
      externalTrail.forEach((point) => {
        xValues.push(point.pos_x);
        yValues.push(point.pos_y);
      });

      const minX = Math.min(...xValues);
      const maxX = Math.max(...xValues);
      const minY = Math.min(...yValues);
      const maxY = Math.max(...yValues);
      const targetX = normalize(currentRobot.pos_x, minX, maxX, 0.18);
      const targetY = normalize(currentRobot.pos_y, minY, maxY, 0.22);
      const animated = animatedRef.current;

      animated.x = lerp(animated.x, targetX, 0.08);
      animated.y = lerp(animated.y, targetY, 0.08);
      animated.heading = lerpAngle(animated.heading, currentRobot.heading, 0.12);

      const liveTrail = trailRef.current;
      liveTrail.push({ x: animated.x, y: animated.y });
      if (liveTrail.length > 48) {
        liveTrail.splice(0, liveTrail.length - 48);
      }

      const renderedTrail = externalTrail.length > 1
        ? externalTrail.map((point) => ({
            x: normalize(point.pos_x, minX, maxX, 0.18),
            y: normalize(point.pos_y, minY, maxY, 0.22),
          }))
        : liveTrail;

      context.clearRect(0, 0, width, height);
      drawSolarField(context, width, height, phase);

      context.strokeStyle = 'rgba(23,119,255,0.22)';
      context.lineWidth = 3;
      context.beginPath();
      renderedTrail.forEach((point, index) => {
        const px = point.x * width;
        const py = point.y * height;
        if (index === 0) {
          context.moveTo(px, py);
        } else {
          context.lineTo(px, py);
        }
      });
      context.stroke();

      currentPeers
        .filter((item) => item.robot_id !== currentRobot.robot_id)
        .slice(0, 6)
        .forEach((item) => {
          const px = normalize(item.pos_x, minX, maxX, 0.18) * width;
          const py = normalize(item.pos_y, minY, maxY, 0.22) * height;
          context.fillStyle = item.online_status === 1 ? 'rgba(55,199,118,0.3)' : 'rgba(160,160,160,0.24)';
          context.beginPath();
          context.arc(px, py, 5, 0, Math.PI * 2);
          context.fill();
        });

      const robotX = animated.x * width;
      const robotY = animated.y * height;
      const statusColor = STATUS_COLORS[currentRobot.work_status] || STATUS_COLORS[0];
      const glowColor = currentRobot.online_status === 1 ? ONLINE_GLOW : OFFLINE_GLOW;
      const speedFactor = clamp(currentRobot.speed / 3, 0.15, 1);
      const energy = clamp(currentRobot.battery_level / 100, 0, 1);

      context.save();
      context.translate(robotX, robotY);
      context.rotate((animated.heading * Math.PI) / 180);

      context.fillStyle = `rgba(${currentRobot.online_status === 1 ? '55,199,118' : '180,180,180'},0.18)`;
      context.beginPath();
      context.ellipse(0, 26, 28, 9, 0, 0, Math.PI * 2);
      context.fill();

      context.strokeStyle = glowColor;
      context.lineWidth = 2;
      context.shadowColor = glowColor;
      context.shadowBlur = 14;
      context.beginPath();
      context.arc(0, 0, 30 + Math.sin(phase * 4) * 2, 0, Math.PI * 2);
      context.stroke();
      context.shadowBlur = 0;

      drawRobotBody(context, currentRobot.robot_type, statusColor, energy);

      if (currentRobot.work_status === 1) {
        context.fillStyle = 'rgba(64,209,255,0.28)';
        context.beginPath();
        context.moveTo(16, -18);
        context.lineTo(58 + speedFactor * 14, -8);
        context.lineTo(58 + speedFactor * 14, 8);
        context.lineTo(16, 18);
        context.closePath();
        context.fill();

        context.strokeStyle = 'rgba(255,255,255,0.55)';
        context.lineWidth = 2;
        for (let i = 0; i < 3; i += 1) {
          const offset = (phase * 42 + i * 12) % 36;
          context.beginPath();
          context.moveTo(16 + offset, -12);
          context.lineTo(6 + offset, 12);
          context.stroke();
        }
      }

      if (currentRobot.work_status === 2) {
        context.strokeStyle = 'rgba(255,182,72,0.65)';
        context.lineWidth = 3;
        context.beginPath();
        context.arc(0, 0, 36 + Math.sin(phase * 6) * 4, 0, Math.PI * 2);
        context.stroke();
      }

      if (currentRobot.work_status === 3) {
        context.fillStyle = 'rgba(255,90,95,0.22)';
        context.fillRect(-40, -40, 80, 80);
      }

      if (currentRobot.work_status === 4) {
        context.strokeStyle = 'rgba(135,88,255,0.8)';
        context.setLineDash([5, 5]);
        context.strokeRect(-34, -28, 68, 56);
        context.setLineDash([]);
      }

      context.restore();

      context.fillStyle = 'rgba(15,23,42,0.82)';
      context.font = '600 14px sans-serif';
      context.fillText(currentRobot.robot_name, 16, 24);

      context.fillStyle = 'rgba(15,23,42,0.64)';
      context.font = '12px sans-serif';
      context.fillText(`速度 ${currentRobot.speed.toFixed(2)} m/s`, 16, 44);
      context.fillText(`清扫 ${currentRobot.clean_area.toFixed(2)} m²`, 16, 62);
      context.fillText(`电量 ${currentRobot.battery_level}%`, 16, 80);

      context.fillStyle = 'rgba(255,255,255,0.86)';
      context.fillRect(width - 142, 16, 126, 14);
      context.fillStyle = '#d9e6f3';
      context.fillRect(width - 142, 16, 126, 14);
      context.fillStyle = currentRobot.battery_level < 20 ? '#ff5a5f' : currentRobot.battery_level < 50 ? '#faad14' : '#37c776';
      context.fillRect(width - 142, 16, 126 * energy, 14);

      const latestFooterLabel = latestFooterLabelRef.current;
      if (latestFooterLabel) {
        context.fillStyle = 'rgba(15,23,42,0.62)';
        context.font = '12px sans-serif';
        context.fillText(latestFooterLabel, 16, height - 16);
      }

      frameId = window.requestAnimationFrame(render);
    };

    frameId = window.requestAnimationFrame(render);

    return () => {
      window.cancelAnimationFrame(frameId);
      observer.disconnect();
    };
  }, []);

  return (
    <div
      style={{
        borderRadius: 14,
        overflow: 'hidden',
        border: '1px solid #dbe7f3',
        background: 'linear-gradient(180deg, #f8fcff 0%, #eef6fb 100%)',
        ...style,
      }}
    >
      <canvas
        ref={canvasRef}
        style={{ width: '100%', height: 260, display: 'block' }}
      />
    </div>
  );
}
