import Link from 'next/link'
import { Container } from '@/components/Container'
import { ProductGrid } from '@/components/ProductGrid'
import { SearchBar } from '@/components/SearchBar'
import { api, type Product } from '@/lib/api'

// Отключаем кеширование - всегда загружаем свежие данные с бэка
export const revalidate = 0

export default async function ShopAllPage(props: { searchParams: Promise<{ search?: string }> }) {
  const { search } = await props.searchParams
  let products: Product[] = []
  let error: string | null = null

  try {
    if (search) {
      const data = await api.searchProducts(search)
      products = data.items || []
    } else {
      const data = await api.getProducts({ limit: 50 })
      products = data.items || []
    }
  } catch (e) {
    error = e instanceof Error ? e.message : 'Ошибка загрузки товаров'
    console.error('Failed to load products:', e)
  }

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">магазин</div>
        <h1 className="h2">Смотреть все</h1>
        <SearchBar />
        {error && <p className="lead">{error}</p>}
        {search && products.length === 0 && (
          <p className="lead">По запросу &quot;{search}&quot; ничего не найдено</p>
        )}
        <div style={{ marginTop: '1rem', display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '2rem' }}>
          <Link href="/shop/new">Новые поступления</Link>
          <Link href="/shop/sale">Сезонные скидки</Link>
        </div>
        {products.length > 0 && <ProductGrid products={products} />}
      </Container>
    </section>
  )
}
