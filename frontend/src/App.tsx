import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useAuthStore } from './store/authStore';
import AppLayout from './components/Layout/AppLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Monitor from './pages/Monitor';
import Stations from './pages/Stations';
import Robots from './pages/Robots';
import Tasks from './pages/Tasks';
import Alarms from './pages/Alarms';
import Analytics from './pages/Analytics';
import Predictions from './pages/Predictions';
import Firmwares from './pages/Firmwares';
import Maintenance from './pages/Maintenance';
import Cameras from './pages/Cameras';
import System from './pages/System';
import Profile from './pages/Profile';
import Geofences from './pages/Geofences';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: '#1677ff',
          colorBgLayout: '#f0f5ff',
          borderRadius: 6,
        },
        components: {
          Menu: {
            itemSelectedBg: '#e6f4ff',
            itemSelectedColor: '#1677ff',
          },
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              path="/"
              element={
                <PrivateRoute>
                  <AppLayout />
                </PrivateRoute>
              }
            >
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="monitor" element={<Monitor />} />
              <Route path="stations" element={<Stations />} />
              <Route path="robots" element={<Robots />} />
              <Route path="tasks" element={<Tasks />} />
              <Route path="alarms" element={<Alarms />} />
              <Route path="analytics" element={<Analytics />} />
              <Route path="predictions" element={<Predictions />} />
              <Route path="firmwares" element={<Firmwares />} />
              <Route path="maintenance" element={<Maintenance />} />
              <Route path="cameras" element={<Cameras />} />
              <Route path="geofences" element={<Geofences />} />
              <Route path="system/*" element={<System />} />
              <Route path="profile" element={<Profile />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
}
