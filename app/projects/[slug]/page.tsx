import { Container } from '@/components/Container'

export default async function ProjectPage(props: { params: Promise<{ slug: string }> }) {
  const { slug } = await props.params

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">проект</div>
        <h1 className="h2">{slug}</h1>
        <p className="lead">
          TODO: грузим проект по slug из <code>/api/projects/{'{slug}'}</code>. Здесь будет текст, галерея и мета.
        </p>
      </Container>
    </section>
  )
}

