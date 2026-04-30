import { ReloadOutlined } from "@ant-design/icons";
import { Button, Col, Descriptions, Form, Input, Row, Space, Table, Tabs, Tag, Timeline } from "antd";
import type { TableColumnsType } from "antd";
import { HealthTag } from "../components/HealthTag";
import { SectionCard } from "../components/SectionCard";
import { backendPhases, dictionaryEvents, dictionaryProperties, ingestionRules } from "../domain/mockData";
import type { DictionaryEvent, DictionaryProperty, IngestionRule } from "../domain/types";
import { useConsoleStore } from "../store/consoleStore";

export function SettingsPage() {
  const site = useConsoleStore((store) => store.site);
  const allowedDomains = useConsoleStore((store) => store.allowedDomains);
  const addAllowedDomain = useConsoleStore((store) => store.addAllowedDomain);
  const reset = useConsoleStore((store) => store.reset);

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={10}>
          <SectionCard title="Site configuration">
            <Form layout="vertical" initialValues={site}>
              <Form.Item label="Site name" name="name">
                <Input />
              </Form.Item>
              <Form.Item label="Primary domain" name="domain">
                <Input />
              </Form.Item>
              <Form.Item label="Website id">
                <Input value={site.id} readOnly />
              </Form.Item>
            </Form>
            <Space wrap>
              {allowedDomains.map((domain) => (
                <Tag key={domain}>{domain}</Tag>
              ))}
            </Space>
            <Space className="form-actions">
              <Button onClick={() => addAllowedDomain(`app.${site.domain}`)}>Add app domain</Button>
              <Button icon={<ReloadOutlined />} onClick={reset}>
                Reset local state
              </Button>
            </Space>
          </SectionCard>
        </Col>
        <Col xs={24} xl={14}>
          <SectionCard title="Data contract summary">
            <Descriptions bordered size="small" column={{ xs: 1, md: 3 }}>
              <Descriptions.Item label="Events">{dictionaryEvents.length}</Descriptions.Item>
              <Descriptions.Item label="Properties">{dictionaryProperties.length}</Descriptions.Item>
              <Descriptions.Item label="Rules">{ingestionRules.length}</Descriptions.Item>
              <Descriptions.Item label="Nested JSON">rejected</Descriptions.Item>
              <Descriptions.Item label="PII">blocked</Descriptions.Item>
              <Descriptions.Item label="Unknown events">quarantine</Descriptions.Item>
            </Descriptions>
          </SectionCard>
        </Col>
      </Row>
      <SectionCard title="Governance">
        <Tabs
          items={[
            {
              key: "events",
              label: "Events",
              children: <Table rowKey="name" columns={eventColumns} dataSource={dictionaryEvents} size="small" pagination={false} />,
            },
            {
              key: "properties",
              label: "Properties",
              children: <Table rowKey="key" columns={propertyColumns} dataSource={dictionaryProperties} size="small" pagination={false} />,
            },
            {
              key: "rules",
              label: "Ingestion rules",
              children: <Table rowKey="rule" columns={ruleColumns} dataSource={ingestionRules} size="small" pagination={false} />,
            },
            {
              key: "roadmap",
              label: "Phases",
              children: (
                <Timeline
                  items={backendPhases.map((phase) => ({
                    color: phase.phase === "P1" ? "green" : "gray",
                    children: (
                      <Space direction="vertical" size={0}>
                        <strong>
                          {phase.phase} · {phase.name}
                        </strong>
                        <span className="muted-text">{phase.scope}</span>
                      </Space>
                    ),
                  }))}
                />
              ),
            },
          ]}
        />
      </SectionCard>
    </Space>
  );
}

const eventColumns: TableColumnsType<DictionaryEvent> = [
  { title: "Event", dataIndex: "name" },
  { title: "Status", dataIndex: "status", render: (value: DictionaryEvent["status"]) => <HealthTag value={value} /> },
  { title: "Required properties", dataIndex: "required" },
];

const propertyColumns: TableColumnsType<DictionaryProperty> = [
  { title: "Property", dataIndex: "key" },
  { title: "Type", dataIndex: "type" },
  { title: "Allowed values", dataIndex: "values" },
];

const ruleColumns: TableColumnsType<IngestionRule> = [
  { title: "Rule", dataIndex: "rule" },
  { title: "Detail", dataIndex: "detail" },
  { title: "Mode", dataIndex: "mode", render: (value: IngestionRule["mode"]) => <HealthTag value={value === "enforced" ? "enforced" : "review"} /> },
];
