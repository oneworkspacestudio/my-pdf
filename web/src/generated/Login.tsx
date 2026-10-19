// Generated from screens/Login.aui; edit the AUI source.
import { Alert, Button, Card, Field, Heading, Image, Input, Link, Page, Text } from "@/components/ui"

export function Login({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Card variant="login">
        <Image src={data.logo} alt={data.name} width={80} height={80} />
        <Heading level={1}>{data.name}</Heading>
        <Text>{data.loginText}</Text>
        {data.error && (
          <Alert tone="error">{data.error}</Alert>
        )}
        {data.oidc ? (
          <Link href={data.loginURL} variant="button">Continue with MyAccounts</Link>
        ) : (
          <>
            <Field label="Email address"><Input name="email" type="email" autoComplete="username" /></Field>
            <Field label="Password"><Input name="password" type="password" autoComplete="current-password" /></Field>
            <Button onClick={() => onAction("login")}>{data.loginButton}</Button>
          </>
        )}
        <Link href={data.workspace}>One Workspace</Link>
      </Card>
    </Page>
  )
}
