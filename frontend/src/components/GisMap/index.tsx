import { useEffect, useRef } from 'react';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import robotFixedImg from '../../assets/robot-fixed.png';
import robotFerryImg from '../../assets/robot-ferry-correct.png';
import robotSmartImg from '../../assets/robot-smart.png';

delete (L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
});

interface StationMarker {
  station_id: string;
  station_name: string;
  longitude: number;
  latitude: number;
}

interface RobotMarker {
  robot_id: string;
  robot_name: string;
  robot_type: number;
  pos_x: number;   // longitude
  pos_y: number;   // latitude
  heading: number;
  online_status: number;
  work_status: number;
  battery_level: number;
}

interface GisMapProps {
  stations: StationMarker[];
  robots: RobotMarker[];
  center?: [number, number];
  zoom?: number;
  selectedRobotId?: string;
  selectedRobotIds?: Set<string>;
  onRobotSelect?: (robotId: string, options?: { additive?: boolean }) => void;
  style?: React.CSSProperties;
}

const WORK_STATUS_COLORS: Record<number, string> = {
  0: '#8c8c8c',
  1: '#1890ff',
  2: '#faad14',
  3: '#ff4d4f',
  4: '#722ed1',
};

// ---- image cache ----
const robotImages: Record<number, HTMLImageElement> = {};
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

// ---- Canvas overlay implementation ----

class RobotCanvasLayer {
  private canvas: HTMLCanvasElement | null = null;
  private ctx: CanvasRenderingContext2D | null = null;
  private map: L.Map;
  private drawFrame: number | null = null;

  robots: RobotMarker[] = [];
  selectedRobotId?: string;
  selectedRobotIds?: Set<string>;
  onRobotSelect?: (robotId: string, options?: { additive?: boolean }) => void;

  constructor(map: L.Map) {
    this.map = map;
  }

  getCanvas(): HTMLCanvasElement | null {
    return this.canvas;
  }

  attach(container: HTMLElement) {
    const canvas = L.DomUtil.create('canvas', 'robot-canvas-layer') as HTMLCanvasElement;
    canvas.style.position = 'absolute';
    canvas.style.pointerEvents = 'auto';
    canvas.style.zIndex = '400';
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');

    this.map.on('move zoom resize moveend zoomend', this.resize.bind(this));
    this.resize();

    canvas.addEventListener('click', (e: MouseEvent) => {
      if (!this.onRobotSelect) return;
      const hit = this.hitTest(e.offsetX, e.offsetY);
      if (hit) {
        this.onRobotSelect(hit, { additive: e.ctrlKey || e.metaKey });
      }
    });

    container.appendChild(canvas);
    this.scheduleDraw();
  }

  remove() {
    if (this.drawFrame != null) cancelAnimationFrame(this.drawFrame);
    this.drawFrame = null;
    this.canvas?.remove();
    this.canvas = null;
    this.ctx = null;
  }

  private resize() {
    if (!this.canvas) return;
    const size = this.map.getSize();
    const dpr = window.devicePixelRatio || 1;
    this.canvas.style.width = `${size.x}px`;
    this.canvas.style.height = `${size.y}px`;
    this.canvas.width = Math.round(size.x * dpr);
    this.canvas.height = Math.round(size.y * dpr);
    if (this.ctx) this.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    this.scheduleDraw();
  }

  scheduleDraw() {
    if (this.drawFrame != null) return;
    this.drawFrame = requestAnimationFrame(() => {
      this.drawFrame = null;
      this.draw();
    });
  }

  private draw() {
    const ctx = this.ctx;
    const canvas = this.canvas;
    if (!ctx || !canvas) return;
    const map = this.map;
    const zoom = map.getZoom();
    const cw = canvas.clientWidth;
    const ch = canvas.clientHeight;

    ctx.clearRect(0, 0, cw, ch);

    const selId = this.selectedRobotId;
    const selIds = this.selectedRobotIds;

    for (const r of this.robots) {
      const pt = map.latLngToContainerPoint([r.pos_y, r.pos_x]);
      const sx = pt.x;
      const sy = pt.y;

      // skip off-screen
      if (sx < -20 || sy < -20 || sx > cw + 20 || sy > ch + 20) continue;

      const isSelected = r.robot_id === selId;
      const isBatchSelected = selIds?.has(r.robot_id) ?? false;
      const color = r.online_status === 0 ? '#d9d9d9' : (WORK_STATUS_COLORS[r.work_status] || '#8c8c8c');
      const angle = (r.heading * Math.PI) / 180;

      // LOD based on zoom — use PNG images
      const img = robotImages[r.robot_type];
      const hasImg = img && img.complete && img.naturalWidth > 0;

      if (zoom < 13) {
        // Small dot or tiny image
        if (hasImg) {
          ctx.save();
          ctx.translate(sx, sy);
          ctx.rotate(angle);
          ctx.drawImage(img, -6, -6, 12, 12);
          ctx.restore();
        } else {
          ctx.beginPath();
          ctx.arc(sx, sy, 3, 0, Math.PI * 2);
          ctx.fillStyle = color;
          ctx.fill();
        }
        if (isSelected || isBatchSelected) {
          ctx.beginPath();
          ctx.arc(sx, sy, 5, 0, Math.PI * 2);
          ctx.strokeStyle = isSelected ? '#1683ff' : '#f3a11b';
          ctx.lineWidth = 2;
          ctx.stroke();
        }
      } else if (zoom < 16) {
        // Medium image with rotation
        if (hasImg) {
          ctx.save();
          ctx.translate(sx, sy);
          ctx.rotate(angle);
          ctx.drawImage(img, -10, -10, 20, 20);
          ctx.restore();
        } else {
          ctx.save();
          ctx.translate(sx, sy);
          ctx.rotate(angle);
          ctx.beginPath();
          ctx.arc(0, 0, 5, 0, Math.PI * 2);
          ctx.fillStyle = color;
          ctx.fill();
          ctx.fillStyle = '#fff';
          ctx.beginPath();
          ctx.moveTo(4, 0);
          ctx.lineTo(-2, -3);
          ctx.lineTo(0, 0);
          ctx.lineTo(-2, 3);
          ctx.closePath();
          ctx.fill();
          ctx.restore();
        }
        if (isSelected || isBatchSelected) {
          ctx.beginPath();
          ctx.arc(sx, sy, 8, 0, Math.PI * 2);
          ctx.strokeStyle = isSelected ? '#1683ff' : '#f3a11b';
          ctx.lineWidth = 2;
          ctx.stroke();
        }
      } else {
        // Large image with name + battery
        if (hasImg) {
          ctx.save();
          ctx.translate(sx, sy);
          ctx.rotate(angle);
          ctx.drawImage(img, -18, -18, 36, 36);
          ctx.restore();
        } else {
          ctx.save();
          ctx.translate(sx, sy);
          ctx.rotate(angle);
          ctx.beginPath();
          ctx.arc(0, 0, 6, 0, Math.PI * 2);
          ctx.fillStyle = color;
          ctx.fill();
          ctx.fillStyle = '#fff';
          ctx.beginPath();
          ctx.moveTo(5, 0);
          ctx.lineTo(-3, -4);
          ctx.lineTo(0, 0);
          ctx.lineTo(-3, 4);
          ctx.closePath();
          ctx.fill();
          ctx.restore();
        }

        if (isSelected || isBatchSelected) {
          ctx.beginPath();
          ctx.arc(sx, sy, 10, 0, Math.PI * 2);
          ctx.strokeStyle = isSelected ? '#1683ff' : '#f3a11b';
          ctx.lineWidth = isSelected ? 3 : 2;
          ctx.stroke();
        }

        // name
        ctx.fillStyle = 'rgba(15,23,42,0.82)';
        ctx.font = isSelected ? '600 12px sans-serif' : '11px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText(r.robot_name || r.robot_id, sx, sy - 14);

        // battery bar
        const barW = 22;
        const barH = 3;
        ctx.fillStyle = 'rgba(0,0,0,0.15)';
        ctx.fillRect(sx - barW / 2, sy + 12, barW, barH);
        ctx.fillStyle = r.battery_level < 20 ? '#ff4d4f' : r.battery_level < 50 ? '#faad14' : '#52c41a';
        ctx.fillRect(sx - barW / 2, sy + 12, (r.battery_level / 100) * barW, barH);
      }
    }

    // Draw offline in grey, alarm (work_status 3) in red priority is already handled by color
  }

  private hitTest(sx: number, sy: number): string | null {
    const map = this.map;
    const zoom = map.getZoom();
    const threshold = zoom < 13 ? 8 : zoom < 16 ? 10 : 14;

    let bestId: string | null = null;
    let bestDist = Infinity;

    for (const r of this.robots) {
      const pt = map.latLngToContainerPoint([r.pos_y, r.pos_x]);
      const d = Math.hypot(pt.x - sx, pt.y - sy);
      if (d < threshold && d < bestDist) {
        bestDist = d;
        bestId = r.robot_id;
      }
    }

    return bestId;
  }
}

// ---- React component ----

export default function GisMap({
  stations,
  robots,
  center = [39.9, 116.4],
  zoom = 12,
  selectedRobotId,
  selectedRobotIds,
  onRobotSelect,
  style,
}: GisMapProps) {
  const mapRef = useRef<L.Map | null>(null);
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const stationMarkersRef = useRef<Map<string, L.Marker>>(new Map());
  const robotLayerRef = useRef<RobotCanvasLayer | null>(null);

  // ---- Leaflet map init ----
  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return;

    mapRef.current = L.map(mapContainerRef.current).setView(center, zoom);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(mapRef.current);

    // Create robot canvas layer in overlayPane
    const map = mapRef.current;
    const overlayPane = map.getPanes()?.overlayPane;
    if (overlayPane) {
      const layer = new RobotCanvasLayer(map);
      layer.attach(overlayPane);
      robotLayerRef.current = layer;
    }

    return () => {
      robotLayerRef.current?.remove();
      robotLayerRef.current = null;
      mapRef.current?.remove();
      mapRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ---- Keep Leaflet view in sync with selected station ----
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    map.setView(center, zoom);
    window.requestAnimationFrame(() => {
      map.invalidateSize();
      robotLayerRef.current?.scheduleDraw();
    });
  }, [center, zoom]);

  // ---- Update robot layer data ----
  useEffect(() => {
    const layer = robotLayerRef.current;
    if (!layer) return;
    layer.robots = robots;
    layer.selectedRobotId = selectedRobotId;
    layer.selectedRobotIds = selectedRobotIds;
    layer.onRobotSelect = onRobotSelect;
    layer.scheduleDraw();
  }, [robots, selectedRobotId, selectedRobotIds, onRobotSelect]);

  // ---- Update station markers ----
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const existing = stationMarkersRef.current;
    const newIds = new Set(stations.map((s) => s.station_id));

    existing.forEach((marker, id) => {
      if (!newIds.has(id)) { marker.remove(); existing.delete(id); }
    });

    stations.forEach((s) => {
      const existingMarker = existing.get(s.station_id);
      if (existingMarker) {
        existingMarker.setLatLng([s.latitude, s.longitude]);
      } else {
        const marker = L.marker([s.latitude, s.longitude])
          .bindPopup(`<b>${s.station_name}</b>`)
          .addTo(map);
        existing.set(s.station_id, marker);
      }
    });
  }, [stations]);

  return <div ref={mapContainerRef} style={{ width: '100%', height: '100%', ...style }} />;
}
