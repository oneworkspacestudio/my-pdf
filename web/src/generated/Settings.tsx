// Generated from screens/Settings.aui; edit the AUI source.
import { Button, Card, Field, Heading, Input, Link, Page, Text } from "@/components/ui"

export function Settings({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Card>
        <Heading level={1}>Settings & backup</Heading>
        <Text>Encrypted backups include this app's database and stored files. Restore into a new, empty NTFS data folder using the executable's --restore command. Existing data is never overwritten.</Text>
        <Field label="Backup password (8–128 bytes)"><Input name="backupPassword" type="password" /></Field>
        <Button onClick={() => onAction("backup")}>Download encrypted backup</Button>
        <Text variant="caption">Keep this password separately. It cannot be recovered. Central Workspace backups also include this app when installed.</Text>
      </Card>
      {data.localOwner && (
        <Card>
          <Heading level={2}>Add a standalone user</Heading>
          <Field label="User email"><Input name="memberEmail" type="email" /></Field>
          <Field label="Initial password"><Input name="memberPassword" type="password" /></Field>
          <Button onClick={() => onAction("member")}>Create user</Button>
        </Card>
      )}
      <Card>
        <Text>{data.versionText}</Text>
        <Text>{data.identityText}</Text>
        {data.appLinks.map((item, i) => (
          <Link key={i} href={item.href}>{item.title}</Link>
        ))}
      </Card>
    </Page>
  )
}
