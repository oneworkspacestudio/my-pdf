// Generated from screens/Shell.aui; edit the AUI source.
import { Alert, Button, Card, Heading, Icon, Image, Link, Page, Row, Section, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Shell({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="sidebar">
        <Row variant="brand">
          <Image src={data.logo} alt={data.name} width={38} height={38} />
          <Stack gap="xs">
            <Heading level={2}>{data.name}</Heading>
            <Text variant="caption">ONE WORKSPACE</Text>
          </Stack>
        </Row>
        {data.nav.map((item, i) => (
          <Link key={i} href={item.href} variant={item.variant}><Text>{item.title}</Text></Link>
        ))}
        <Spacer />
        <Text variant="caption">Your data stays on your server.</Text>
        <Link href="#settings" variant="nav">Settings & backup</Link>
        {data.workspace && (
          <Link href={data.workspace} variant="nav">← One Workspace</Link>
        )}
      </Stack>
      <Stack variant="workspace">
        <Row variant="topbar">
          <Heading level={2}>{data.heading}</Heading>
          <Spacer />
          {data.searchURL && (
            <Link href={data.searchURL}>Search Workspace</Link>
          )}
          <Button variant="icon" onClick={() => onAction("theme")} aria-label="Toggle color theme"><Icon name="SunMoon" /></Button>
          <Button variant="secondary" onClick={() => onAction("apps")}>Apps</Button>
          <Text>{data.user.name}</Text>
          <Button variant="secondary" onClick={() => onAction("logout")}>Sign out</Button>
        </Row>
        <Stack variant="content">
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert tone="success">{data.notice}</Alert>
          )}
          {data.busy && (
            <Row>
              <Spinner />
              <Text>{data.busy}</Text>
            </Row>
          )}
          <Section id="screen-outlet" />
        </Stack>
        {data.appsOpen && (
          <Card variant="dialog" aria-label="Workspace apps">
            <Heading level={2}>Your Workspace apps</Heading>
            {data.appLinks.map((item, i) => (
              <Link key={i} href={item.href}>{item.title}</Link>
            ))}
            <Button variant="secondary" onClick={() => onAction("closeApps")}>Close</Button>
          </Card>
        )}
      </Stack>
    </Page>
  )
}
