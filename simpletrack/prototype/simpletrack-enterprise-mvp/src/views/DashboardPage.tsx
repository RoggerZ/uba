import { PauseCircleOutlined, PlayCircleOutlined } from "@ant-design/icons";
import { Line } from "@ant-design/charts";
import { Button, Col, Row, Segmented, Space, Statistic, Table, Tag, Timeline, Typography } from "antd";
import type { TableColumnsType } from "antd";
import { useEffect, useMemo, useState } from "react";
import { HealthTag } from "../components/HealthTag";
import { SectionCard } from "../components/SectionCard";
import { kpis, topPages, topReferrers, traffic } from "../domain/mockData";
import type { LiveSignal } from "../domain/types";
import { useConsoleStore } from "../store/consoleStore";
import { integer } from "../utils/format";

export function DashboardPage() {
  const [range, setRange] = useState<string | number>("7d");
  const livePaused = useConsoleStore((store) => store.livePaused);
  const liveSignals = useConsoleStore((store) => store.liveSignals);
  const setLivePaused = useConsoleStore((store) => store.setLivePaused);
  const pushLiveSignal = useConsoleStore((store) => store.pushLiveSignal);

  useEffect(() => {
    const timer = window.setInterval(pushLiveSignal, 7000);
    return () => window.clearInterval(timer);
  }, [pushLiveSignal]);

  const chartData = useMemo(
    () =>
      traffic.flatMap((point) => [
        { date: point.date, metric: "Pageviews", value: point.pageviews },
        { date: point.date, metric: "Visitors", value: point.visitors },
        { date: point.date, metric: "Events", value: point.events },
      ]),
    [],
  );

  const lineConfig = {
    data: chartData,
    xField: "date",
    yField: "value",
    colorField: "metric",
    height: 260,
    autoFit: true,
    axis: { y: { title: false }, x: { title: false } },
    legend: { position: "bottom" as const },
  };

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <div className="toolbar-row">
        <Segmented value={range} onChange={setRange} options={["24h", "7d", "30d"]} />
        <Button icon={livePaused ? <PlayCircleOutlined /> : <PauseCircleOutlined />} onClick={() => setLivePaused(!livePaused)}>
          {livePaused ? "Resume live" : "Pause live"}
        </Button>
      </div>
      <Row gutter={[12, 12]}>
        {kpis.map((metric) => (
          <Col xs={24} sm={12} xl={6} key={metric.label}>
            <SectionCard title={metric.label}>
              <Statistic value={metric.value} suffix={<span className={metric.trend === "up" ? "delta-up" : "delta-down"}>{metric.delta}</span>} />
            </SectionCard>
          </Col>
        ))}
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={15}>
          <SectionCard title="Traffic trend" extra={<Tag>{range}</Tag>}>
            <Line {...lineConfig} />
          </SectionCard>
        </Col>
        <Col xs={24} xl={9}>
          <SectionCard title="Realtime intake" extra={<HealthTag value={livePaused ? "draft" : "healthy"} />}>
            <Timeline
              items={liveSignals.map((signal) => ({
                color: signal.status === "accepted" ? "green" : "orange",
                children: <SignalItem signal={signal} />,
              }))}
            />
          </SectionCard>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={12}>
          <SectionCard title="Top pages">
            <Table columns={pageColumns} dataSource={topPages} size="small" pagination={false} />
          </SectionCard>
        </Col>
        <Col xs={24} xl={12}>
          <SectionCard title="Top referrers">
            <Table columns={referrerColumns} dataSource={topReferrers} size="small" pagination={false} />
          </SectionCard>
        </Col>
      </Row>
    </Space>
  );
}

function SignalItem({ signal }: { signal: LiveSignal }) {
  return (
    <Space direction="vertical" size={0}>
      <Typography.Text strong>{signal.name}</Typography.Text>
      <Typography.Text type="secondary">
        {signal.time} · {signal.path} · {signal.visitor}
      </Typography.Text>
    </Space>
  );
}

const pageColumns: TableColumnsType<(typeof topPages)[number]> = [
  { title: "Path", dataIndex: "path" },
  { title: "Views", dataIndex: "views", align: "right", render: (value: number) => integer.format(value) },
  { title: "Visitors", dataIndex: "visitors", align: "right", render: (value: number) => integer.format(value) },
];

const referrerColumns: TableColumnsType<(typeof topReferrers)[number]> = [
  { title: "Source", dataIndex: "source" },
  { title: "Visitors", dataIndex: "visitors", align: "right", render: (value: number) => integer.format(value) },
  { title: "Share", dataIndex: "share", align: "right" },
];
