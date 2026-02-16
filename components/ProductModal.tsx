'use client'

import { useState, useEffect } from 'react'
import Image from 'next/image'
import { addToCart } from '@/lib/cart'
import { api, getPlaceholderImage, getImageUrl, formatPrice, type Product } from '@/lib/api'
import styles from './ProductModal.module.css'

type ProductModalProps = {
  product: Product | null
  isOpen: boolean
  onClose: () => void
}

export function ProductModal({ product, isOpen, onClose }: ProductModalProps) {
  const [currentImageIndex, setCurrentImageIndex] = useState(0)
  const [adding, setAdding] = useState(false)
  const [imageError, setImageError] = useState(false)

  useEffect(() => {
    if (product) {
      setCurrentImageIndex(0)
      setImageError(false)
    }
  }, [product])

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [isOpen])

  if (!isOpen || !product) return null

  const images = product.images && product.images.length > 0 
    ? product.images.map(img => getImageUrl(img))
    : [getPlaceholderImage(600, 600)]
  const currentImage = images[currentImageIndex] || images[0]
  const price = formatPrice(product.price, product.currency)

  const handlePrevImage = () => {
    setCurrentImageIndex((prev) => (prev === 0 ? images.length - 1 : prev - 1))
    setImageError(false)
  }

  const handleNextImage = () => {
    setCurrentImageIndex((prev) => (prev === images.length - 1 ? 0 : prev + 1))
    setImageError(false)
  }

  const handleAddToCart = () => {
    setAdding(true)
    addToCart({
      productId: product.id,
      slug: product.slug,
      title: product.title,
      price: product.price,
      currency: product.currency,
      image: currentImage,
    })
    window.dispatchEvent(new Event('cartUpdated'))
    setTimeout(() => {
      setAdding(false)
      onClose()
    }, 300)
  }

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div className={styles.modal} onClick={handleBackdropClick}>
      <div className={styles.modalContent}>
        <button className={styles.closeButton} onClick={onClose} aria-label="Закрыть">
          ×
        </button>

        <div className={styles.productGrid}>
          <div className={styles.imageSection}>
            {images.length > 1 && (
              <button className={styles.navButton} onClick={handlePrevImage} aria-label="Предыдущее фото">
                ‹
              </button>
            )}
            
            <div className={styles.imageContainer}>
              {!imageError ? (
                <Image
                  src={currentImage}
                  alt={product.title}
                  fill
                  sizes="(max-width: 768px) 100vw, 50vw"
                  className={styles.image}
                  unoptimized={currentImage.includes('picsum.photos')}
                  onError={() => setImageError(true)}
                />
              ) : (
                <div className={styles.imagePlaceholder}>
                  <span>📷</span>
                </div>
              )}
            </div>

            {images.length > 1 && (
              <button className={styles.navButton} onClick={handleNextImage} aria-label="Следующее фото">
                ›
              </button>
            )}

            {images.length > 1 && (
              <div className={styles.imageIndicators}>
                {images.map((_, index) => (
                  <button
                    key={index}
                    className={`${styles.indicator} ${index === currentImageIndex ? styles.active : ''}`}
                    onClick={() => {
                      setCurrentImageIndex(index)
                      setImageError(false)
                    }}
                    aria-label={`Фото ${index + 1}`}
                  />
                ))}
              </div>
            )}
          </div>

          <div className={styles.details}>
            <div className={styles.badges}>
              {product.isNew && <span className={styles.badge}>NEW</span>}
              {product.isOnSale && <span className={styles.badgeSale}>SALE</span>}
            </div>

            <h2 className={styles.title}>{product.title}</h2>
            
            {product.description && (
              <p className={styles.description}>{product.description}</p>
            )}

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
      </div>
    </div>
  )
}
