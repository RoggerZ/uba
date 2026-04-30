import { Column } from "@ant-design/charts";
import { Input, Row, Col, Segmented, Space, Table, Typography } from "antd";
import type { TableColumnsType } from "antd";
import { useMemo, useState } from "react";
import { HealthTag } from "../components/HealthTag";
import { SectionCard } from "../components/SectionCard";
import { analyticsEvents } from "../domain/mockData";
import type { AnalyticsEvent } from "../domain/types";
import { useConsoleStore } from "../store/consoleStore";
import { integer, propertyRows } from "../utils/format";

export function EventsPage() {
  const [search, setSearch] = useState("");
  const selectedEventKey = useConsoleStore((store) => store.selectedEventKey);
  const selectedProperty = useConsoleStore((store) => store.selectedProperty);
  const selectEvent = useConsoleStore((store) => store.selectEvent);
  const selectProperty = useConsoleStore((store) => store.selectProperty);

  const filteredEvents = analyticsEvents.filter((event) => event.name.includes(search.trim().toLowerCase()));
  const selected = analyticsEvents.find((event) => event.key === selectedEventKey) ?? analyticsEvents[0];
  const propertyKeys = Object.keys(selected.properties);
  const activeProperty = propertyKeys.includes(selectedProperty) ? selectedProperty : propertyKeys[0];
  const distribution = propertyRows(selected.properties[activeProperty]);

  const columns = useMemo<TableColumnsType<AnalyticsEvent>>(
    () => [
      {
        title: "Event",
        dataIndex: "name",
        render: (_, record) => (
          <Space direction="vertical" size={0}>
            <Typography.Text strong>{record.name}</Typography.Text>
            <Typography.Text type="secondary">{record.description}</Typography.Text>
          </Space>
        ),
      },
      { title: "Count", dataIndex: "count", align: "right", render: (value: number) => integer.format(value) },
      { title: "Visitors", dataIndex: "visitors", align: "right", render: (value: number) => integer.format(value) },
      { title: "Last seen", dataIndex: "lastSeen" },
      { title: "Health", dataIndex: "health", render: (value: AnalyticsEvent["health"]) => <HealthTag value={value} /> },
    ],
    [],
  );

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <div className="toolbar-row">
        <Input.Search placeholder="Search event name" allowClear value={search} onChange={(event) => setSearch(event.target.value)} />
      </div>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={15}>
          <SectionCard title="Event dictionary">
            <Table
              rowKey="key"
              columns={columns}
              dataSource={filteredEvents}
              size="small"
              pagination={false}
              rowClassName={(record) => (record.key === selected.key ? "selected-row" : "")}
              onRow={(record) => ({
                onClick: () => selectEvent(record.key, Object.keys(record.properties)[0]),
              })}
            />
          </SectionCard>
        </Col>
        <Col xs={24} xl={9}>
          <SectionCard title={selected.name} extra={`${integer.format(selected.count)} events`}>
            <Space direction="vertical" size={14} className="full-width">
              <Segmented block value={activeProperty} options={propertyKeys} onChange={(value) => selectProperty(String(value))} />
              <Column
                data={distribution}
                xField="name"
                yField="value"
                height={280}
                colorField="name"
                legend={false}
                axis={{ x: { title: false }, y: { title: false } }}
              />
            </Space>
          </SectionCard>
        </Col>
      </Row>
    </Space>
  );
}
