'use client'

import { useState, useEffect } from 'react'
import { Container } from '@/components/Container'
import { GalleryGrid } from '@/components/GalleryGrid'
import { api, type GalleryItem } from '@/lib/api'

const CATEGORIES = [
  { label: 'Все', value: '' },
  { label: 'Интро', value: 'intro' },
] as const

// Это клиентский компонент, revalidate не нужен

export default function GalleryPage() {
  const [category, setCategory] = useState<string>('')
  const [items, setItems] = useState<GalleryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadGallery() {
      setLoading(true)
      setError(null)
      try {
        const response = await api.getGalleryItems(category || undefined)
        setItems(response.items || [])
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка загрузки галереи')
        console.error('Failed to load gallery:', e)
      } finally {
        setLoading(false)
      }
    }

    loadGallery()
  }, [category])

  return (
    <section className="section">
      <Container size="wide">
        <div className="kicker">галерея</div>
        <h1 className="h2">Смотреть</h1>

        <div style={{ marginTop: '1.25rem', display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginBottom: '2rem' }}>
          {CATEGORIES.map((cat) => (
            <button
              key={cat.value}
              onClick={() => setCategory(cat.value)}
              style={{
                border: category === cat.value ? '1px solid var(--fg)' : '1px solid var(--line)',
                padding: '0.35rem 0.6rem',
                borderRadius: 999,
                fontSize: '0.9rem',
                color: category === cat.value ? 'var(--fg)' : 'var(--muted)',
                background: 'transparent',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {cat.label}
            </button>
          ))}
        </div>

        {error ? (
          <div style={{ color: 'var(--muted)', padding: '2rem 0' }}>{error}</div>
        ) : (
          <GalleryGrid items={items} loading={loading} />
        )}
      </Container>
    </section>
  )
}
