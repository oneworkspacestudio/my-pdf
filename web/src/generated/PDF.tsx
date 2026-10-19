// Generated from screens/PDF.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Grid, Heading, Image, Input, Link, Page, Row, Section, Select, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function PDF({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row>
        <Heading level={1}>Your PDF workspace</Heading>
        <Spacer />
        <Button onClick={() => onAction("upload")}>Upload PDF / images</Button>
        <Button variant="secondary" onClick={() => onAction("drive")}>Open from Drive</Button>
      </Row>
      <Field label="Choose PDF, PNG or JPEG files (PDF limit 100 MiB; workspace storage 1 GiB per account)"><Input name="uploads" type="file" multiple={true} accept="application/pdf,image/png,image/jpeg" /></Field>
      <Alert>Original files are preserved. Annotations and drawing a signature do not create a cryptographically signed PDF. Processed copies can invalidate existing signatures.</Alert>
      <Text>{data.storageText}</Text>
      {data.managed && (
        <Checkbox name="autoDrive">Save new uploads and processed PDFs to Drive automatically</Checkbox>
      )}
      {data.saveStatus && (
        <Alert>{data.saveStatus}</Alert>
      )}
      {data.exporting && (
        <Button variant="secondary" onClick={() => onAction("cancelExport")} cancel={true}>Cancel browser export</Button>
      )}
      {data.viewing && (
        <Card>
          <Row variant="pdf-toolbar">
            <Button variant="secondary" onClick={() => onAction("previous")}>← Previous</Button>
            <Text>{data.pageLabel}</Text>
            <Button variant="secondary" onClick={() => onAction("next")}>Next →</Button>
            <Select name="zoom" options={data.zooms} />
            <Button variant="secondary" onClick={() => onAction("rotateView")}>Rotate view</Button>
            <Button variant="secondary" onClick={() => onAction("print")}>Print / browser viewer</Button>
          </Row>
          <Row>
            <Input name="pageNumber" type="number" min={1} max={data.pageCount} />
            <Button variant="secondary" onClick={() => onAction("goPage")}>Go to page</Button>
          </Row>
          <Row variant="thumbnails">
            {data.thumbnails.map((item, i) => (
              <Button key={i} variant="thumbnail" item={item.page} onClick={() => onAction("match")}>
                <Image src={item.src} alt={item.label} width={54} />
                <Text>{item.label}</Text>
              </Button>
            ))}
          </Row>
          <Section id="pdf-canvas" />
          <Row>
            <Input name="find" placeholder="Find text in this PDF" />
            <Button variant="secondary" onClick={() => onAction("find")}>Find</Button>
            <Button variant="secondary" onClick={() => onAction("extractText")}>Extract text</Button>
            <Button variant="secondary" onClick={() => onAction("image")}>Export page PNG</Button>
          </Row>
          <Row>
            <Select name="tool" options={data.annotationTools} />
            <Input name="annotationText" placeholder="Text / comment / stamp text" />
            <Button variant="secondary" onClick={() => onAction("undoDrawing")}>Undo last mark on page</Button>
            <Button variant="secondary" onClick={() => onAction("clearDrawings")}>Clear page marks</Button>
          </Row>
          <Field label="Stamp / signature image (PNG or JPEG)"><Input name="stampUpload" type="file" accept="image/png,image/jpeg" /></Field>
          <Button variant="secondary" onClick={() => onAction("stampImage")}>Use stamp image</Button>
          <Grid columns={4}>
            <Field label="Mark X (0–1)"><Input name="markX" step={0.01} type="number" min={0} max={1} /></Field>
            <Field label="Mark Y (0–1)"><Input name="markY" step={0.01} type="number" min={0} max={1} /></Field>
            <Field label="Mark width (0–1)"><Input name="markW" step={0.01} type="number" min={0} max={1} /></Field>
            <Field label="Mark height (0–1)"><Input name="markH" step={0.01} type="number" min={0} max={1} /></Field>
          </Grid>
          <Button variant="secondary" onClick={() => onAction("addMark")}>Add mark using coordinates</Button>
          <Text variant="caption">Drag to draw a rectangle or ink; click for text/comments. Marks are saved per document and account. Black redaction previews become permanent only after Rasterise & redact.</Text>
          <Row>
            <Button onClick={() => onAction("saveAnnotations")}>Save annotated copy</Button>
            <Button variant="danger" onClick={() => onAction("redact")}>Rasterise & redact</Button>
            <Button variant="secondary" onClick={() => onAction("watermark")}>Watermark copy</Button>
            <Button variant="secondary" onClick={() => onAction("screen")}>Compress: screen-quality copy</Button>
            <Button variant="secondary" onClick={() => onAction("numbers")}>Number pages</Button>
            <Button variant="secondary" onClick={() => onAction("metadata")}>Remove metadata copy</Button>
            <Button variant="secondary" onClick={() => onAction("saveDrive")}>Save selected copy to Drive</Button>
          </Row>
          <Field label="Watermark text"><Input name="watermark" /></Field>
          {data.textOutput && (
            <Text variant="pre">{data.textOutput}</Text>
          )}
          {data.matches.map((item, i) => (
            <Button key={i} variant="secondary" item={item.page} onClick={() => onAction("match")}>{item.label}</Button>
          ))}
        </Card>
      )}
      <Grid columns={2}>
        <Card>
          <Heading level={2}>Documents</Heading>
          {data.documents.map((item, i) => (
            <Row key={i}>
              <Button variant="secondary" item={item.id} onClick={() => onAction("select")}>{item.selection}</Button>
              <Link href={item.href}>{item.name}</Link>
              <Text variant="caption">{item.sizeLabel}</Text>
              <Text variant="caption">{item.driveLabel}</Text>
              <Button variant="secondary" item={item.id} onClick={() => onAction("open")}>Open</Button>
              {item.canSaveDrive && (
                <Button variant="secondary" item={item.id} onClick={() => onAction("queueDrive")}>Save to Drive</Button>
              )}
              {item.canRetryDrive && (
                <Button variant="secondary" item={item.id} onClick={() => onAction("retryDrive")}>Retry Drive save</Button>
              )}
              <Button variant="secondary" item={item.id} onClick={() => onAction("removeDocument")}>Remove</Button>
            </Row>
          ))}
        </Card>
        <Card>
          <Heading level={2}>Tools</Heading>
          <Select name="operation" options={data.operations} />
          <Field label="Pages (e.g. 1-3,5; reorder: 3,1,2)"><Input name="pages" /></Field>
          <Field label="Rotation degrees"><Select name="rotation" options={data.rotations} /></Field>
          <Field label="PDF password (used in memory for this operation only)"><Input name="pdfPassword" type="password" /></Field>
          <Button onClick={() => onAction("process")}>Process selected PDFs</Button>
          <Text variant="caption">Annotated exports are flattened raster copies. Compression is lossless structural optimisation; scanned-image quality is preserved. OCR and Office conversion are not included.</Text>
        </Card>
      </Grid>
      {data.driveOpen && (
        <Card>
          <Heading level={2}>My Drive picker</Heading>
          <Field label="Your Drive source"><Select name="driveSource" options={data.driveSources} /></Field>
          <Field label="Folder (inside your allowed scope)"><Input name="drivePath" placeholder="/" /></Field>
          <Button variant="secondary" onClick={() => onAction("listDrive")}>List folder</Button>
          {data.driveFiles.map((item, i) => (
            <Row key={i}>
              <Text>{item.name}</Text>
              <Button variant="secondary" item={item.path} onClick={() => onAction("openDrive")}>{item.action}</Button>
            </Row>
          ))}
          <Field label="New copy filename"><Input name="driveName" placeholder="edited-document.pdf" /></Field>
          <Text variant="caption">Saving always creates a new file. Existing files are never overwritten.</Text>
          <Button onClick={() => onAction("confirmSaveDrive")}>Confirm: save a new copy</Button>
          <Button variant="secondary" onClick={() => onAction("closeDrive")}>Close picker</Button>
        </Card>
      )}
      <Card>
        <Heading level={2}>Jobs</Heading>
        {data.jobs.map((item, i) => (
          <Row key={i}>
            <Stack gap="xs">
              <Text>{item.operation}</Text>
              <Text variant="caption">{item.message}</Text>
            </Stack>
            <Badge>{item.state}</Badge>
            <Text variant="caption">{item.elapsed}</Text>
            {item.active && (
              <>
                <Spinner />
                <Button variant="secondary" item={item.id} onClick={() => onAction("cancel")}>Cancel</Button>
              </>
            )}
            {item.openResult && (
              <Button variant="secondary" item={item.result} onClick={() => onAction("open")}>Open result</Button>
            )}
            {item.downloadURL && (
              <>
                <Link href={item.downloadURL}>Download split pages ZIP</Link>
                <Button variant="secondary" item={item.result} onClick={() => onAction("removeArtifact")}>Remove ZIP</Button>
              </>
            )}
            {item.metadataText && (
              <Text variant="pre">{item.metadataText}</Text>
            )}
          </Row>
        ))}
      </Card>
    </Page>
  )
}
