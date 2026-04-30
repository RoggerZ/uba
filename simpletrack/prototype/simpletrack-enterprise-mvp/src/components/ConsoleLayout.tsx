import {
  ApiOutlined,
  BarChartOutlined,
  CheckCircleOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  FlagOutlined,
  SettingOutlined,
} from "@ant-design/icons";
import { Layout, Menu, Space, Tag, Typography } from "antd";
import type { MenuProps } from "antd";
import { Link, Outlet, useLocation } from "react-router-dom";
import { backendPhases } from "../domain/mockData";
import { useConsoleStore } from "../store/consoleStore";

const { Header, Sider, Content } = Layout;

const items: MenuProps["items"] = [
  { key: "/onboarding", icon: <CheckCircleOutlined />, label: <Link to="/onboarding">Onboarding</Link> },
  { key: "/dashboard", icon: <DashboardOutlined />, label: <Link to="/dashboard">Dashboard</Link> },
  { key: "/events", icon: <DatabaseOutlined />, label: <Link to="/events">Events</Link> },
  { key: "/goals", icon: <FlagOutlined />, label: <Link to="/goals">Goals</Link> },
  { key: "/settings", icon: <SettingOutlined />, label: <Link to="/settings">Settings</Link> },
];

export function ConsoleLayout() {
  const location = useLocation();
  const site = useConsoleStore((store) => store.site);
  const selectedKey = `/${location.pathname.split("/").filter(Boolean)[0] || "onboarding"}`;

  return (
    <Layout className="console-shell">
      <Sider width={260} breakpoint="lg" className="console-sider">
        <div className="brand">
          <div className="brand-mark">ST</div>
          <div>
            <Typography.Text strong>SimpleTrack</Typography.Text>
            <Typography.Text type="secondary">Production Console</Typography.Text>
          </div>
        </div>
        <Menu mode="inline" selectedKeys={[selectedKey]} items={items} />
        <div className="phase-list">
          {backendPhases.map((phase) => (
            <div className={phase.phase === "P1" ? "phase-item is-current" : "phase-item"} key={phase.phase}>
              <span>{phase.phase}</span>
              <strong>{phase.name}</strong>
            </div>
          ))}
        </div>
      </Sider>
      <Layout>
        <Header className="console-header">
          <Space direction="vertical" size={0}>
            <Typography.Text type="secondary">{site.environment}</Typography.Text>
            <Typography.Title level={3}>{pageTitle(selectedKey)}</Typography.Title>
          </Space>
          <Space wrap>
            <Tag icon={<ApiOutlined />}>{site.id}</Tag>
            <Tag icon={<BarChartOutlined />}>{site.name}</Tag>
            <Tag color="success">Ingest healthy</Tag>
          </Space>
        </Header>
        <Content className="console-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

function pageTitle(route: string) {
  const titles: Record<string, string> = {
    "/onboarding": "Connect and verify the first signal",
    "/dashboard": "Site health and live intake",
    "/events": "Event contract and property evidence",
    "/goals": "Simple goal definitions",
    "/settings": "Site settings and data rules",
  };
  return titles[route] ?? "SimpleTrack Console";
}
