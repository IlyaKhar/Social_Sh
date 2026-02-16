'use client'

import { useState } from 'react'
import Image from 'next/image'
import { getPlaceholderImage, getImageUrl, type GalleryItem } from '@/lib/api'
import styles from './GalleryGrid.module.css'

type GalleryGridProps = {
  items: GalleryItem[]
  loading?: boolean
}

export function GalleryGrid({ items, loading }: GalleryGridProps) {
  if (loading) {
    return (
      <div className={styles.grid}>
        {[...Array(6)].map((_, i) => (
          <div key={i} className={styles.skeleton} />
        ))}
      </div>
    )
  }

  if (items.length === 0) {
    return <div className={styles.empty}>Фотографии не найдены</div>
  }

  return (
    <div className={styles.grid}>
      {items.map((item) => {
        const imageUrl = item.image ? getImageUrl(item.image) : getPlaceholderImage(600, 600)
        return <GalleryItem key={item.id} item={item} imageUrl={imageUrl} />
      })}
    </div>
  )
}

function GalleryItem({ item, imageUrl }: { item: GalleryItem; imageUrl: string }) {
  const [imageError, setImageError] = useState(false)

  return (
    <div className={styles.item}>
      {!imageError ? (
        <Image
          src={imageUrl}
          alt={item.title || 'Галерея'}
          fill
          sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
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
  )
}
