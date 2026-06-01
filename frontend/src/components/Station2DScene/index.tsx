import { useEffect, useRef, type CSSProperties } from 'react';
import robotFixedImg from '../../assets/robot-fixed.png';
import robotFerryImg from '../../assets/robot-ferry-correct.png';
import robotSmartImg from '../../assets/robot-smart.png';
import pvBgImg from '../../assets/pv-bg.jpg';

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
  speed?: number;
  clean_area?: number;
}

interface Station2DSceneProps {
  stationName?: string;
  robots: SceneRobot[];
  selectedRobotId?: string;
  selectedRobotIds?: Set<string>;
  onRobotSelect?: (robotId: string, options?: { additive?: boolean }) => void;
  onSelectionChange?: (robotIds: string[], options?: { additive?: boolean }) => void;
  style?: CSSProperties;
}

interface Point {
  x: number;
  y: number;
}

interface Viewport {
  zoom: number;
  panX: number;
  panY: number;
}

interface Rect {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
}

const WORLD_BOUNDS = { minX: 0, maxX: 120, minY: 0, maxY: 67 };
const ZOOM_MIN = 0.4;
const ZOOM_MAX = 20;
// 面板阵列参数对齐 pv-bg.jpg 图片布局 (1376×768 ≈ 1.79:1)
// 5 行光伏板，每行 6 列，行间距中包含接驳车通道
const PANEL_COLS = 6;
const PANEL_ROWS = 5;
const PANEL_CELL_W = 17;
const PANEL_CELL_H = 11;
const PANEL_ORIGIN_X = 9;
const PANEL_ORIGIN_Y = 5;
const PANEL_GAP_Y = 13.6; // 行间距（面板顶部到下一行面板顶部）

// ---- image cache (preload on module level) ----
const robotImages: Record<number, HTMLImageElement> = {};
const bgImage = new Image();
bgImage.src = pvBgImg;
(function preload() {
  const srcMap: Record<number, string> = {
    1: robotFixedImg,
    2: robotFerryImg,
    3: robotSmartImg,
  };
  for (const [type, src] of Object.entries(srcMap)) {
    const img = new Image();
    img.src = src;
    robotImages[Number(type)] = img;
  }
})();

// ---- 清扫轨迹追踪：记录每块光伏板被机器人清扫过的程度 ----
const cleaningTrail = new Map<string, number>(); // key: "col,row" → 0..1
const CLEAN_DECAY = 0.9995; // 清扫痕迹渐隐速度
const CLEAN_SPEED = 0.015;  // 机器人经过时清扫进度增速

function getPanelCell(wx: number, wy: number): { col: number; row: number } | null {
  const col = Math.floor((wx - PANEL_ORIGIN_X) / PANEL_CELL_W);
  const row = Math.floor((wy - PANEL_ORIGIN_Y) / PANEL_GAP_Y);
  if (col < 0 || col >= PANEL_COLS || row < 0 || row >= PANEL_ROWS) return null;
  return { col, row };
}

function updateCleaningTrail(robots: SceneRobot[], dt: number) {
  // 渐隐所有清扫痕迹
  for (const [key, val] of cleaningTrail) {
    const newVal = val * Math.pow(CLEAN_DECAY, dt * 60);
    if (newVal < 0.01) cleaningTrail.delete(key);
    else cleaningTrail.set(key, newVal);
  }
  // 机器人当前位置增加清扫进度
  for (const r of robots) {
    if (r.work_status !== 1 || r.online_status !== 1) continue;
    // 机器人周围的面板都标记为清扫中
    for (let dx = -3; dx <= 3; dx += 3) {
      const cell = getPanelCell(r.pos_x + dx, r.pos_y);
      if (cell) {
        const key = `${cell.col},${cell.row}`;
        const cur = cleaningTrail.get(key) || 0;
        cleaningTrail.set(key, Math.min(1, cur + CLEAN_SPEED));
      }
    }
  }
}

function getCleaningLevel(col: number, row: number): number {
  return cleaningTrail.get(`${col},${row}`) || 0;
}

const STATUS_COLORS: Record<number, string> = {
  0: '#7d8792',
  1: '#1683ff',
  2: '#f3a11b',
  3: '#e23d4f',
  4: '#7452d9',
};

// ---- coordinate helpers ----

function clampZoom(z: number): number {
  return Math.max(ZOOM_MIN, Math.min(ZOOM_MAX, z));
}

function worldToScreen(wx: number, wy: number, vp: Viewport, cw: number, ch: number): Point {
  return {
    x: (wx + vp.panX) * vp.zoom + cw / 2,
    y: (wy + vp.panY) * vp.zoom + ch / 2,
  };
}

function screenToWorld(sx: number, sy: number, vp: Viewport, cw: number, ch: number): Point {
  return {
    x: (sx - cw / 2) / vp.zoom - vp.panX,
    y: (sy - ch / 2) / vp.zoom - vp.panY,
  };
}

function getVisibleWorldRect(vp: Viewport, cw: number, ch: number): Rect {
  const tl = screenToWorld(0, 0, vp, cw, ch);
  const br = screenToWorld(cw, ch, vp, cw, ch);
  return {
    minX: tl.x - 8,
    minY: tl.y - 8,
    maxX: br.x + 8,
    maxY: br.y + 8,
  };
}

// ---- drawing helpers ----

function drawRobotLow(
  ctx: CanvasRenderingContext2D,
  r: SceneRobot,
  sx: number,
  sy: number,
  selected: boolean,
  batchSelected: boolean,
) {
  const img = robotImages[r.robot_type];
  if (img && img.complete && img.naturalWidth > 0) {
    ctx.save();
    ctx.translate(sx, sy);
    const angle = (r.heading * Math.PI) / 180;
    ctx.rotate(angle);
    const size = 12;
    ctx.drawImage(img, -size / 2, -size / 2, size, size);
    ctx.restore();
  } else {
    const color = r.online_status === 1 ? (STATUS_COLORS[r.work_status] || STATUS_COLORS[0]) : '#aeb6bf';
    ctx.beginPath();
    ctx.arc(sx, sy, selected || batchSelected ? 6 : 4, 0, Math.PI * 2);
    ctx.fillStyle = color;
    ctx.fill();
  }
  if (selected || batchSelected) {
    ctx.strokeStyle = selected ? '#1683ff' : '#f3a11b';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(sx, sy, selected || batchSelected ? 6 : 4, 0, Math.PI * 2);
    ctx.stroke();
  }
}

function drawRobotMid(
  ctx: CanvasRenderingContext2D,
  r: SceneRobot,
  sx: number,
  sy: number,
  selected: boolean,
  batchSelected: boolean,
) {
  const img = robotImages[r.robot_type];
  const angle = (r.heading * Math.PI) / 180;

  ctx.save();
  ctx.translate(sx, sy);
  ctx.rotate(angle);

  // shadow
  ctx.fillStyle = selected ? 'rgba(22,131,255,0.16)' : 'rgba(15,23,42,0.1)';
  ctx.beginPath();
  ctx.ellipse(0, 16, 22, 6, 0, 0, Math.PI * 2);
  ctx.fill();

  // robot image
  if (img && img.complete && img.naturalWidth > 0) {
    const size = 28;
    ctx.drawImage(img, -size / 2, -size / 2, size, size);
  } else {
    const color = r.online_status === 1 ? (STATUS_COLORS[r.work_status] || STATUS_COLORS[0]) : '#aeb6bf';
    ctx.fillStyle = color;
    ctx.beginPath();
    ctx.arc(0, 0, 14, 0, Math.PI * 2);
    ctx.fill();
  }

  // offline overlay
  if (r.online_status === 0) {
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.beginPath();
    ctx.arc(0, 0, 16, 0, Math.PI * 2);
    ctx.fill();
  }

  ctx.restore();

  // selection ring
  if (selected || batchSelected) {
    ctx.beginPath();
    ctx.arc(sx, sy, 20, 0, Math.PI * 2);
    ctx.strokeStyle = selected ? '#1683ff' : '#f3a11b';
    ctx.lineWidth = selected ? 3 : 2;
    ctx.stroke();
  }

  // name
  ctx.fillStyle = 'rgba(15,23,42,0.74)';
  ctx.font = selected ? '600 11px sans-serif' : '10px sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText(r.robot_name || r.robot_id, sx, sy - 22);
}

function drawRobotHigh(
  ctx: CanvasRenderingContext2D,
  r: SceneRobot,
  sx: number,
  sy: number,
  selected: boolean,
  batchSelected: boolean,
) {
  const img = robotImages[r.robot_type];
  const angle = (r.heading * Math.PI) / 180;

  ctx.save();
  ctx.translate(sx, sy);
  ctx.rotate(angle);

  // shadow
  ctx.fillStyle = selected ? 'rgba(22,131,255,0.16)' : 'rgba(15,23,42,0.1)';
  ctx.beginPath();
  ctx.ellipse(0, 22, 28, 8, 0, 0, Math.PI * 2);
  ctx.fill();

  // robot image
  if (img && img.complete && img.naturalWidth > 0) {
    const size = 46;
    ctx.drawImage(img, -size / 2, -size / 2, size, size);
  } else {
    const color = r.online_status === 1 ? (STATUS_COLORS[r.work_status] || STATUS_COLORS[0]) : '#aeb6bf';
    ctx.fillStyle = color;
    ctx.beginPath();
    ctx.arc(0, 0, 18, 0, Math.PI * 2);
    ctx.fill();
  }

  // offline overlay
  if (r.online_status === 0) {
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.beginPath();
    ctx.arc(0, 0, 24, 0, Math.PI * 2);
    ctx.fill();
  }

  // work status sweep
  if (r.work_status === 1 && r.online_status === 1) {
    ctx.fillStyle = 'rgba(40,190,255,0.24)';
    ctx.beginPath();
    ctx.moveTo(16, -15);
    ctx.lineTo(52, -7);
    ctx.lineTo(52, 7);
    ctx.lineTo(16, 15);
    ctx.closePath();
    ctx.fill();
  }

  ctx.restore();

  // selection ring
  if (selected || batchSelected) {
    ctx.beginPath();
    ctx.arc(sx, sy, 27, 0, Math.PI * 2);
    ctx.strokeStyle = selected ? '#1683ff' : '#f3a11b';
    ctx.lineWidth = selected ? 3 : 2;
    ctx.stroke();
  }

  // name
  ctx.fillStyle = 'rgba(15,23,42,0.74)';
  ctx.font = selected ? '600 13px sans-serif' : '12px sans-serif';
  ctx.textAlign = 'center';
  ctx.fillText(r.robot_name || r.robot_id, sx, sy - 34);

  // battery bar
  const barW = 36;
  const barH = 4;
  const barX = sx - barW / 2;
  const barY = sy + 28;
  ctx.fillStyle = 'rgba(0,0,0,0.12)';
  ctx.fillRect(barX, barY, barW, barH);
  const fillW = (r.battery_level / 100) * barW;
  ctx.fillStyle = r.battery_level < 20 ? '#ff4d4f' : r.battery_level < 50 ? '#faad14' : '#52c41a';
  ctx.fillRect(barX, barY, fillW, barH);
}

// ---- selection rect ----

function drawSelectionRect(
  ctx: CanvasRenderingContext2D,
  start: Point,
  end: Point,
) {
  const x = Math.min(start.x, end.x);
  const y = Math.min(start.y, end.y);
  const w = Math.abs(end.x - start.x);
  const h = Math.abs(end.y - start.y);
  ctx.strokeStyle = '#1683ff';
  ctx.lineWidth = 1;
  ctx.setLineDash([4, 3]);
  ctx.strokeRect(x, y, w, h);
  ctx.fillStyle = 'rgba(22,131,255,0.08)';
  ctx.fillRect(x, y, w, h);
  ctx.setLineDash([]);
}

// ---- main component ----

export default function Station2DScene({
  stationName,
  robots,
  selectedRobotId,
  selectedRobotIds,
  onRobotSelect,
  onSelectionChange,
  style,
}: Station2DSceneProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const robotsRef = useRef(robots);
  const selectedRobotIdRef = useRef(selectedRobotId);
  const selectedRobotIdsRef = useRef(selectedRobotIds);
  const onRobotSelectRef = useRef(onRobotSelect);
  const onSelectionChangeRef = useRef(onSelectionChange);
  const viewportRef = useRef<Viewport>({ zoom: 10, panX: -60, panY: -33.5 });
  const visibleRobotsRef = useRef<Array<{ robotId: string; sx: number; sy: number; robot: SceneRobot }>>([]);
  const pointerStateRef = useRef<{
    mode: 'idle' | 'panning' | 'selecting' | 'dragging-robot';
    downPoint?: Point;
    currentPoint?: Point;
    startViewport?: Viewport;
    downTime?: number;
  }>({ mode: 'idle' });
  const lastRenderTimeRef = useRef(0);

  useEffect(() => { robotsRef.current = robots; }, [robots]);
  useEffect(() => { selectedRobotIdRef.current = selectedRobotId; }, [selectedRobotId]);
  useEffect(() => { selectedRobotIdsRef.current = selectedRobotIds; }, [selectedRobotIds]);
  useEffect(() => { onRobotSelectRef.current = onRobotSelect; }, [onRobotSelect]);
  useEffect(() => { onSelectionChangeRef.current = onSelectionChange; }, [onSelectionChange]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return undefined;
    const ctx = canvas.getContext('2d');
    if (!ctx) return undefined;

    let frameId = 0;

    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;
      canvas.width = Math.max(1, Math.round(rect.width * dpr));
      canvas.height = Math.max(1, Math.round(rect.height * dpr));
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    resize();
    const observer = new ResizeObserver(resize);
    observer.observe(canvas);

    const render = (time: number) => {
      const cw = canvas.clientWidth;
      const ch = canvas.clientHeight;
      const vp = viewportRef.current;
      const currentRobots = robotsRef.current;
      const selId = selectedRobotIdRef.current;
      const selIds = selectedRobotIdsRef.current;

      // frame throttle
      const frameInterval = currentRobots.length > 100 ? 1000 / 15 : 1000 / 30;
      if (time - lastRenderTimeRef.current < frameInterval) {
        frameId = window.requestAnimationFrame(render);
        return;
      }
      lastRenderTimeRef.current = time;

      // background — pv-bg.jpg 铺满画布
      if (bgImage.complete && bgImage.naturalWidth > 0) {
        ctx.drawImage(bgImage, 0, 0, cw, ch);
      } else {
        const gradient = ctx.createLinearGradient(0, 0, 0, ch);
        gradient.addColorStop(0, '#eaf7ff');
        gradient.addColorStop(1, '#f6f8ed');
        ctx.fillStyle = gradient;
        ctx.fillRect(0, 0, cw, ch);
      }

      // grid lines (world space)
      for (let gx = WORLD_BOUNDS.minX; gx <= WORLD_BOUNDS.maxX; gx += 10) {
        const s = worldToScreen(gx, WORLD_BOUNDS.minY, vp, cw, ch);
        const e = worldToScreen(gx, WORLD_BOUNDS.maxY, vp, cw, ch);
        ctx.strokeStyle = 'rgba(31,76,111,0.06)';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(s.x, s.y);
        ctx.lineTo(e.x, e.y);
        ctx.stroke();
      }
      for (let gy = WORLD_BOUNDS.minY; gy <= WORLD_BOUNDS.maxY; gy += 10) {
        const s = worldToScreen(WORLD_BOUNDS.minX, gy, vp, cw, ch);
        const e = worldToScreen(WORLD_BOUNDS.maxX, gy, vp, cw, ch);
        ctx.strokeStyle = 'rgba(31,76,111,0.06)';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(s.x, s.y);
        ctx.lineTo(e.x, e.y);
        ctx.stroke();
      }

      // panel arrays (world space) — 半透明叠加在背景图之上
      // 面板是机器人清扫的目标区域，图片已展示真实面板，代码绘制透明遮罩
      const visibleRect = getVisibleWorldRect(vp, cw, ch);
      for (let pr = 0; pr < PANEL_ROWS; pr++) {
        for (let pc = 0; pc < PANEL_COLS; pc++) {
          const px = PANEL_ORIGIN_X + pc * PANEL_CELL_W;
          const py = PANEL_ORIGIN_Y + pr * PANEL_GAP_Y;
          if (px + PANEL_CELL_W < visibleRect.minX || px > visibleRect.maxX ||
              py + PANEL_CELL_H < visibleRect.minY || py > visibleRect.maxY) continue;
          const s = worldToScreen(px, py, vp, cw, ch);
          const e = worldToScreen(px + PANEL_CELL_W, py + PANEL_CELL_H, vp, cw, ch);
          const cleaned = getCleaningLevel(pc, pr);

          // 面板底色：半透明深蓝，已清扫区域渐变到绿色
          if (cleaned > 0.01) {
            const r = Math.round(33 * (1 - cleaned) + 34 * cleaned);
            const g = Math.round(75 * (1 - cleaned) + 197 * cleaned);
            const b = Math.round(118 * (1 - cleaned) + 118 * cleaned);
            ctx.fillStyle = `rgba(${r},${g},${b},${0.25 + cleaned * 0.35})`;
          } else {
            ctx.fillStyle = 'rgba(33,75,118,0.18)';
          }
          ctx.fillRect(s.x, s.y, e.x - s.x, e.y - s.y);

          // 面板边框
          ctx.strokeStyle = cleaned > 0.3
            ? `rgba(52,199,89,${0.3 + cleaned * 0.4})`
            : 'rgba(139,211,255,0.35)';
          ctx.lineWidth = 1;
          ctx.strokeRect(s.x, s.y, e.x - s.x, e.y - s.y);

          // 清扫进度条（高缩放时显示）
          if (cleaned > 0.01 && vp.zoom > 2) {
            const barH = Math.max(2, (e.y - s.y) * 0.08);
            const barY = s.y + (e.y - s.y) * 0.85;
            ctx.fillStyle = 'rgba(0,0,0,0.2)';
            ctx.fillRect(s.x + 2, barY, e.x - s.x - 4, barH);
            ctx.fillStyle = '#34c759';
            ctx.fillRect(s.x + 2, barY, (e.x - s.x - 4) * cleaned, barH);
          }
        }
      }

      // 更新清扫轨迹：机器人经过的面板变色
      updateCleaningTrail(currentRobots, frameInterval / 1000);

      // robots
      const visible: Array<{ robotId: string; sx: number; sy: number; robot: SceneRobot }> = [];
      for (const robot of currentRobots) {
        if (robot.pos_x < visibleRect.minX || robot.pos_x > visibleRect.maxX ||
            robot.pos_y < visibleRect.minY || robot.pos_y > visibleRect.maxY) continue;
        const sp = worldToScreen(robot.pos_x, robot.pos_y, vp, cw, ch);
        visible.push({ robotId: robot.robot_id, sx: sp.x, sy: sp.y, robot });

        const isSelected = robot.robot_id === selId;
        const isBatchSelected = selIds?.has(robot.robot_id) ?? false;

        if (vp.zoom < 0.75) {
          drawRobotLow(ctx, robot, sp.x, sp.y, isSelected, isBatchSelected);
        } else if (vp.zoom < 1.5 || currentRobots.length > 120) {
          drawRobotMid(ctx, robot, sp.x, sp.y, isSelected, isBatchSelected);
        } else {
          drawRobotHigh(ctx, robot, sp.x, sp.y, isSelected, isBatchSelected);
        }
      }
      visibleRobotsRef.current = visible;

      // selection rect
      const ps = pointerStateRef.current;
      if (ps.mode === 'selecting' && ps.downPoint && ps.currentPoint) {
        drawSelectionRect(ctx, ps.downPoint, ps.currentPoint);
      }

      // HUD
      ctx.fillStyle = 'rgba(15,23,42,0.82)';
      ctx.font = '600 15px sans-serif';
      ctx.textAlign = 'left';
      ctx.fillText(stationName || '光伏板二维场景', 18, 28);

      ctx.fillStyle = 'rgba(15,23,42,0.6)';
      ctx.font = '12px sans-serif';
      const onlineCount = currentRobots.filter((r) => r.online_status === 1).length;
      ctx.fillText(
        `在线 ${onlineCount}/${currentRobots.length} · 缩放 ${vp.zoom.toFixed(1)}x · 可见 ${visible.length}`,
        18, 48,
      );

      frameId = window.requestAnimationFrame(render);
    };

    frameId = window.requestAnimationFrame(render);
    return () => {
      window.cancelAnimationFrame(frameId);
      observer.disconnect();
    };
  }, [stationName]);

  // ---- pointer handlers ----

  const getEventPoint = (e: React.PointerEvent<HTMLCanvasElement>): Point => ({
    x: e.clientX - e.currentTarget.getBoundingClientRect().left,
    y: e.clientY - e.currentTarget.getBoundingClientRect().top,
  });

  const hitTest = (pt: Point): { robotId: string; sx: number; sy: number } | null => {
    const threshold = viewportRef.current.zoom > 1.5 ? 28 : viewportRef.current.zoom > 0.75 ? 22 : 10;
    const targets = visibleRobotsRef.current
      .map((v) => ({ robotId: v.robotId, sx: v.sx, sy: v.sy, d: Math.hypot(v.sx - pt.x, v.sy - pt.y) }))
      .sort((a, b) => a.d - b.d);
    return targets.length > 0 && targets[0].d <= threshold ? targets[0] : null;
  };

  const handlePointerDown = (e: React.PointerEvent<HTMLCanvasElement>) => {
    const pt = getEventPoint(e);
    const vp = viewportRef.current;
    const canvas = canvasRef.current;
    if (!canvas) return;

    const hit = hitTest(pt);

    if (hit && e.shiftKey && onSelectionChangeRef.current) {
      // Shift-click: start selection rect
      pointerStateRef.current = { mode: 'selecting', downPoint: pt, currentPoint: pt };
    } else if (hit) {
      // Click on robot
      pointerStateRef.current = { mode: 'idle', downPoint: pt, downTime: Date.now() };
    } else {
      // Start pan
      pointerStateRef.current = {
        mode: 'panning',
        downPoint: pt,
        currentPoint: pt,
        startViewport: { zoom: vp.zoom, panX: vp.panX, panY: vp.panY },
      };
    }
    canvas.setPointerCapture(e.pointerId);
    e.preventDefault();
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLCanvasElement>) => {
    const ps = pointerStateRef.current;
    if (ps.mode === 'idle') {
      if (ps.downPoint) {
        const pt = getEventPoint(e);
        ps.currentPoint = pt;
        if (Math.hypot(pt.x - ps.downPoint.x, pt.y - ps.downPoint.y) > 8) {
          // transition to panning
          const vp = viewportRef.current;
          ps.mode = 'panning';
          ps.startViewport = { zoom: vp.zoom, panX: vp.panX, panY: vp.panY };
        }
      }
      return;
    }
    if (ps.mode === 'panning' && ps.downPoint && ps.startViewport) {
      const pt = getEventPoint(e);
      const canvas = canvasRef.current;
      if (!canvas) return;
      const dx = (pt.x - ps.downPoint.x) / ps.startViewport.zoom;
      const dy = (pt.y - ps.downPoint.y) / ps.startViewport.zoom;
      viewportRef.current = {
        zoom: ps.startViewport.zoom,
        panX: ps.startViewport.panX + dx,
        panY: ps.startViewport.panY + dy,
      };
    }
    if (ps.mode === 'selecting') {
      ps.currentPoint = getEventPoint(e);
    }
  };

  const handlePointerUp = (e: React.PointerEvent<HTMLCanvasElement>) => {
    const ps = pointerStateRef.current;
    const canvas = canvasRef.current;
    if (!canvas) { ps.mode = 'idle'; return; }

    if (ps.mode === 'selecting' && ps.downPoint && ps.currentPoint && onSelectionChangeRef.current) {
      // compute selection rect in screen coords
      const sx1 = Math.min(ps.downPoint.x, ps.currentPoint.x);
      const sy1 = Math.min(ps.downPoint.y, ps.currentPoint.y);
      const sx2 = Math.max(ps.downPoint.x, ps.currentPoint.x);
      const sy2 = Math.max(ps.downPoint.y, ps.currentPoint.y);

      const ids: string[] = [];
      for (const v of visibleRobotsRef.current) {
        if (v.sx >= sx1 && v.sx <= sx2 && v.sy >= sy1 && v.sy <= sy2) {
          ids.push(v.robotId);
        }
      }
      if (ids.length > 0) {
        onSelectionChangeRef.current(ids, { additive: e.ctrlKey || e.metaKey });
      }
    } else if (ps.mode === 'idle' && ps.downPoint && ps.downTime) {
      // click (no significant movement)
      const elapsed = Date.now() - ps.downTime;
      const disp = ps.currentPoint
        ? Math.hypot(ps.currentPoint.x - ps.downPoint.x, ps.currentPoint.y - ps.downPoint.y)
        : 0;
      if (elapsed < 300 && disp < 8) {
        const hit = hitTest(ps.downPoint);
        if (hit && onRobotSelectRef.current) {
          onRobotSelectRef.current(hit.robotId, { additive: e.ctrlKey || e.metaKey });
        }
      }
    }

    pointerStateRef.current = { mode: 'idle' };
    canvas.releasePointerCapture(e.pointerId);
  };

  const handlePointerCancel = (e: React.PointerEvent<HTMLCanvasElement>) => {
    pointerStateRef.current = { mode: 'idle' };
    canvasRef.current?.releasePointerCapture(e.pointerId);
  };

  const handleWheel = (e: React.WheelEvent<HTMLCanvasElement>) => {
    e.preventDefault();
    const canvas = canvasRef.current;
    if (!canvas) return;
    const cw = canvas.clientWidth;
    const ch = canvas.clientHeight;

    const vp = viewportRef.current;
    const mouseSx = e.clientX - canvas.getBoundingClientRect().left;
    const mouseSy = e.clientY - canvas.getBoundingClientRect().top;
    const worldBefore = screenToWorld(mouseSx, mouseSy, vp, cw, ch);

    const zoomFactor = 1.1;
    const newZoom = clampZoom(e.deltaY < 0 ? vp.zoom * zoomFactor : vp.zoom / zoomFactor);
    const newVp: Viewport = { zoom: newZoom, panX: vp.panX, panY: vp.panY };
    const worldAfter = screenToWorld(mouseSx, mouseSy, newVp, cw, ch);

    viewportRef.current = {
      zoom: newZoom,
      panX: vp.panX + (worldBefore.x - worldAfter.x),
      panY: vp.panY + (worldBefore.y - worldAfter.y),
    };
  };

  const handleDoubleClick = (_e: React.MouseEvent<HTMLCanvasElement>) => {
    // reset view
    viewportRef.current = { zoom: 10, panX: -60, panY: -33.5 };
  };

  return (
    <div style={{ width: '100%', height: '100%', minHeight: 400, background: '#f6fbff', ...style }}>
      <canvas
        ref={canvasRef}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={handlePointerCancel}
        onWheel={handleWheel}
        onDoubleClick={handleDoubleClick}
        style={{
          width: '100%', height: '100%', display: 'block',
          cursor: pointerStateRef.current.mode === 'panning' ? 'grabbing' : 'default',
          touchAction: 'none',
        }}
      />
    </div>
  );
}
