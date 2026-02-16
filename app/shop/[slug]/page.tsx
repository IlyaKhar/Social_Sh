import { Container } from '@/components/Container'
import { ProductDetail } from '@/components/ProductDetail'
import { api } from '@/lib/api'
import { notFound } from 'next/navigation'

export default async function ProductPage(props: { params: Promise<{ slug: string }> }) {
  const { slug } = await props.params

  let product = null
  try {
    const data = await api.getProduct(slug)
    product = data.item
  } catch (e) {
    console.error('Failed to load product:', e)
  }

  if (!product) {
    notFound()
  }

  return (
    <section className="section">
      <Container size="wide">
        <ProductDetail product={product} />
      </Container>
    </section>
  )
}
