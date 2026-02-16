'use client'

import { useState } from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { addToCart } from '@/lib/cart'
import { api, getPlaceholderImage, getImageUrl, formatPrice, type Product } from '@/lib/api'
import styles from './ProductDetail.module.css'

type ProductDetailProps = {
  product: Product
}

export function ProductDetail({ product }: ProductDetailProps) {
  const router = useRouter()
  const [imageError, setImageError] = useState(false)
  const [adding, setAdding] = useState(false)

  const imageUrl = getImageUrl(product.images?.[0])
  const price = formatPrice(product.price, product.currency)

  const handleAddToCart = () => {
    setAdding(true)
    addToCart({
      productId: product.id,
      slug: product.slug,
      title: product.title,
      price: product.price,
      currency: product.currency,
      image: imageUrl,
    })
    window.dispatchEvent(new Event('cartUpdated'))
    setTimeout(() => {
      setAdding(false)
      router.push('/cart')
    }, 300)
  }

  return (
    <div className={styles.product}>
      <div className={styles.imageSection}>
        {!imageError ? (
          <Image
            src={imageUrl}
            alt={product.title}
            width={600}
            height={600}
            className={styles.image}
            unoptimized={imageUrl.includes('picsum.photos')}
            onError={() => setImageError(true)}
          />
        ) : (
          <div className={styles.imagePlaceholder}>
            <span>📷</span>
          </div>
        )}
      </div>

      <div className={styles.details}>
        <div className="kicker">товар</div>
        <h1 className="h2">{product.title}</h1>
        {product.description && <p className="lead">{product.description}</p>}

        <div className={styles.badges}>
          {product.isNew && <span className={styles.badge}>NEW</span>}
          {product.isOnSale && <span className={styles.badgeSale}>SALE</span>}
        </div>

        <div className={styles.price}>{price}</div>

        <button
          onClick={handleAddToCart}
          disabled={adding}
          className={styles.addButton}
        >
          {adding ? 'Добавление...' : 'Добавить в корзину'}
        </button>
      </div>
    </div>
  )
}
