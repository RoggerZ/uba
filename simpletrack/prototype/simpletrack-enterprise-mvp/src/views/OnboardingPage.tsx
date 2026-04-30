import { CheckCircleOutlined, CopyOutlined, DashboardOutlined, DatabaseOutlined } from "@ant-design/icons";
import { Alert, App, Button, Col, Descriptions, Form, Input, Row, Space, Steps, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { SectionCard } from "../components/SectionCard";
import { trackerSnippet } from "../domain/mockData";
import { useConsoleStore } from "../store/consoleStore";

export function OnboardingPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const site = useConsoleStore((store) => store.site);
  const installVerified = useConsoleStore((store) => store.installVerified);
  const setInstallVerified = useConsoleStore((store) => store.setInstallVerified);
  const pushLiveSignal = useConsoleStore((store) => store.pushLiveSignal);

  async function copySnippet() {
    await navigator.clipboard.writeText(trackerSnippet);
    message.success("Tracker snippet copied");
  }

  function verifyData() {
    setInstallVerified(true);
    pushLiveSignal();
    message.success("Incoming signal accepted");
  }

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Alert
        type="info"
        showIcon
        message="P1 review scope"
        description="This flow proves the data pipe is alive: tracker loaded, allowed domain matched, and the first pageview or event appears in realtime."
      />
      <SectionCard title="Connection workflow">
        <Steps
          current={installVerified ? 2 : 1}
          items={[
            { title: "Create site", description: "Name and domain" },
            { title: "Install tracker", description: "Copy snippet" },
            { title: "Verify data", description: "Realtime signal" },
          ]}
        />
      </SectionCard>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={10}>
          <SectionCard title="Site record">
            <Form layout="vertical" initialValues={site}>
              <Form.Item label="Site name" name="name">
                <Input />
              </Form.Item>
              <Form.Item label="Primary domain" name="domain">
                <Input />
              </Form.Item>
              <Form.Item label="Environment" name="environment">
                <Input />
              </Form.Item>
            </Form>
          </SectionCard>
        </Col>
        <Col xs={24} xl={14}>
          <SectionCard title="Tracker snippet" extra={<Button icon={<CopyOutlined />} onClick={copySnippet}>Copy</Button>}>
            <pre className="code-block">{trackerSnippet}</pre>
            <Space wrap>
              <Button type="primary" icon={<CheckCircleOutlined />} onClick={verifyData}>
                Simulate accepted signal
              </Button>
              <Button icon={<DashboardOutlined />} onClick={() => navigate("/dashboard")}>
                Open dashboard
              </Button>
              <Button icon={<DatabaseOutlined />} onClick={() => navigate("/events")}>
                Review events
              </Button>
            </Space>
          </SectionCard>
        </Col>
      </Row>
      <SectionCard title="Verification evidence">
        <Descriptions bordered column={{ xs: 1, md: 2, xl: 4 }} size="small">
          <Descriptions.Item label="Tracker status">
            <Typography.Text type={installVerified ? "success" : "warning"}>
              {installVerified ? "verified" : "waiting"}
            </Typography.Text>
          </Descriptions.Item>
          <Descriptions.Item label="Last accepted signal">{site.lastSeenAt}</Descriptions.Item>
          <Descriptions.Item label="Domain allowlist">{site.domain}</Descriptions.Item>
          <Descriptions.Item label="PII guard">enforced</Descriptions.Item>
          <Descriptions.Item label="Property shape">flat key-value only</Descriptions.Item>
          <Descriptions.Item label="Unknown event policy">quarantine for review</Descriptions.Item>
        </Descriptions>
      </SectionCard>
    </Space>
  );
}
