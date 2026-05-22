import { useEffect, useRef } from 'react';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

// 修正 Leaflet 默认图标路径问题
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
  pos_x: number;
  pos_y: number;
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
  onRobotClick?: (robotId: string) => void;
  style?: React.CSSProperties;
}

const WORK_STATUS_COLORS: Record<number, string> = {
  0: '#8c8c8c', // 空闲 - 灰
  1: '#1890ff', // 清扫中 - 蓝
  2: '#faad14', // 充电中 - 橙
  3: '#ff4d4f', // 故障 - 红
  4: '#722ed1', // 维护 - 紫
};

function createRobotIcon(workStatus: number, online: number, heading: number): L.DivIcon {
  const color = online === 0 ? '#d9d9d9' : (WORK_STATUS_COLORS[workStatus] || '#8c8c8c');
  return L.divIcon({
    html: `<div style="
      width:12px;height:12px;border-radius:50%;background:${color};
      border:2px solid #fff;box-shadow:0 1px 3px rgba(0,0,0,0.3);
      transform:rotate(${heading}deg);
    "><div style="width:0;height:0;border-left:3px solid transparent;border-right:3px solid transparent;border-bottom:8px solid ${color};margin:-6px 0 0 3px;"></div></div>`,
    className: '',
    iconSize: [12, 20],
    iconAnchor: [6, 10],
  });
}

export default function GisMap({
  stations, robots, center = [39.9, 116.4], zoom = 12,
  onRobotClick, style,
}: GisMapProps) {
  const mapRef = useRef<L.Map | null>(null);
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const robotMarkersRef = useRef<Map<string, L.Marker>>(new Map());
  const stationMarkersRef = useRef<Map<string, L.Marker>>(new Map());

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return;

    mapRef.current = L.map(mapContainerRef.current).setView(center, zoom);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(mapRef.current);

    return () => {
      mapRef.current?.remove();
      mapRef.current = null;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 更新电站标记
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const existing = stationMarkersRef.current;
    const newIds = new Set(stations.map((s) => s.station_id));

    // 移除不存在的
    existing.forEach((marker, id) => {
      if (!newIds.has(id)) { marker.remove(); existing.delete(id); }
    });

    // 添加或更新
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

  // 更新机器人标记
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const existing = robotMarkersRef.current;
    const newIds = new Set(robots.map((r) => r.robot_id));

    existing.forEach((marker, id) => {
      if (!newIds.has(id)) { marker.remove(); existing.delete(id); }
    });

    robots.forEach((r) => {
      const icon = createRobotIcon(r.work_status, r.online_status, r.heading);
      const existingMarker = existing.get(r.robot_id);
      if (existingMarker) {
        existingMarker.setLatLng([r.pos_y, r.pos_x]);
        existingMarker.setIcon(icon);
      } else {
        const marker = L.marker([r.pos_y, r.pos_x], { icon })
          .bindPopup(`<b>${r.robot_name}</b><br/>电量: ${r.battery_level}%`)
          .addTo(map);
        marker.on('click', () => onRobotClick?.(r.robot_id));
        existing.set(r.robot_id, marker);
      }
    });
  }, [robots, onRobotClick]);

  return <div ref={mapContainerRef} style={{ width: '100%', height: '100%', ...style }} />;
}
