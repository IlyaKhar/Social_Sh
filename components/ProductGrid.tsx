import { ProductCard } from './ProductCard'
import { type Product } from '@/lib/api'
import styles from './ProductGrid.module.css'

type ProductGridProps = {
  products: Product[]
  loading?: boolean
}

export function ProductGrid({ products, loading }: ProductGridProps) {
  if (loading) {
    return (
      <div className={styles.grid}>
        {[...Array(6)].map((_, i) => (
          <div key={i} className={styles.skeleton} />
        ))}
      </div>
    )
  }

  if (products.length === 0) {
    return <div className={styles.empty}>Товары не найдены</div>
  }

  return (
    <div className={styles.grid}>
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  )
}
