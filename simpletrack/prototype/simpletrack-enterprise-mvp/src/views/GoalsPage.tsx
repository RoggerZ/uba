import { SaveOutlined } from "@ant-design/icons";
import { Button, Col, Descriptions, Form, Input, Row, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useEffect } from "react";
import { HealthTag } from "../components/HealthTag";
import { SectionCard } from "../components/SectionCard";
import type { GoalDefinition } from "../domain/types";
import { useConsoleStore } from "../store/consoleStore";

export function GoalsPage() {
  const [form] = Form.useForm<GoalDefinition>();
  const goals = useConsoleStore((store) => store.goals);
  const selectedGoalId = useConsoleStore((store) => store.selectedGoalId);
  const selectGoal = useConsoleStore((store) => store.selectGoal);
  const updateGoal = useConsoleStore((store) => store.updateGoal);
  const selected = goals.find((goal) => goal.id === selectedGoalId) ?? goals[0];

  useEffect(() => {
    form.setFieldsValue(selected);
  }, [form, selected]);

  function saveGoal(values: GoalDefinition) {
    updateGoal(selected.id, {
      name: values.name,
      rule: values.rule,
      denominator: values.denominator,
    });
  }

  const columns: TableColumnsType<GoalDefinition> = [
    {
      title: "Goal",
      dataIndex: "name",
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <strong>{record.name}</strong>
          <span className="muted-text">{record.denominator}</span>
        </Space>
      ),
    },
    { title: "Rule", dataIndex: "rule" },
    { title: "Rate", dataIndex: "rate", align: "right" },
    { title: "Status", dataIndex: "status", render: (value: GoalDefinition["status"]) => <HealthTag value={value} /> },
  ];

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} xl={14}>
        <SectionCard title="Goal registry" extra={<Tag>P1 simple goals only</Tag>}>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={goals}
            size="small"
            pagination={false}
            rowClassName={(record) => (record.id === selected.id ? "selected-row" : "")}
            onRow={(record) => ({ onClick: () => selectGoal(record.id) })}
          />
        </SectionCard>
      </Col>
      <Col xs={24} xl={10}>
        <SectionCard title="Definition">
          <Form form={form} layout="vertical" onFinish={saveGoal}>
            <Form.Item label="Name" name="name">
              <Input />
            </Form.Item>
            <Form.Item label="Success rule" name="rule">
              <Input />
            </Form.Item>
            <Form.Item label="Denominator" name="denominator">
              <Input />
            </Form.Item>
            <Descriptions bordered size="small" column={1}>
              <Descriptions.Item label="Conversions">{selected.conversions}</Descriptions.Item>
              <Descriptions.Item label="Population">{selected.population}</Descriptions.Item>
              <Descriptions.Item label="Rate">{selected.rate}</Descriptions.Item>
            </Descriptions>
            <Space className="form-actions">
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>
                Save definition
              </Button>
              <Button onClick={() => updateGoal(selected.id, { status: selected.status === "active" ? "draft" : "active" })}>
                {selected.status === "active" ? "Mark draft" : "Activate"}
              </Button>
            </Space>
          </Form>
        </SectionCard>
      </Col>
    </Row>
  );
}
