'use client'

import { useState } from 'react'
import Image from 'next/image'
import { formatPrice, getPlaceholderImage, getImageUrl, type Product } from '@/lib/api'
import { ProductModal } from './ProductModal'
import styles from './ProductCard.module.css'

type ProductCardProps = {
  product: Product
}

export function ProductCard({ product }: ProductCardProps) {
  const [imageError, setImageError] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [imageLoading, setImageLoading] = useState(true)
  const imageUrl = getImageUrl(product.images?.[0] || getPlaceholderImage(400, 500))
  const price = formatPrice(product.price, product.currency)

  return (
    <>
      <article className={styles.card} onClick={() => setIsModalOpen(true)}>
        <div className={styles.imageWrapper}>
          {!imageError ? (
            <>
              {imageLoading && <div className={styles.imageSkeleton} />}
              <Image
                src={imageUrl}
                alt={product.title}
                fill
                sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                className={styles.image}
                loading="lazy"
                unoptimized={imageUrl.includes('picsum.photos')}
                onError={() => setImageError(true)}
                onLoad={() => setImageLoading(false)}
              />
            </>
          ) : (
            <div className={styles.imagePlaceholder}>
              <span>📷</span>
            </div>
          )}
          {product.isNew && <span className={styles.badge}>NEW</span>}
          {product.isOnSale && <span className={styles.badgeSale}>SALE</span>}
        </div>
        <div className={styles.content}>
          <h3 className={styles.title}>{product.title}</h3>
          <p className={styles.price}>{price}</p>
        </div>
      </article>

      <ProductModal
        product={product}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </>
  )
}
