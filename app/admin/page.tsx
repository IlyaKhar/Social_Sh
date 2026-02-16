'use client'

import { useState, useEffect } from 'react'
import { api, type Product, type GalleryItem, type Page } from '@/lib/api'
import styles from './page.module.css'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001'

export default function AdminPage() {
  const [token, setToken] = useState<string | null>(null)
  const [products, setProducts] = useState<Product[]>([])
  const [gallery, setGallery] = useState<GalleryItem[]>([])
  const [pages, setPages] = useState<Page[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'products' | 'gallery' | 'pages'>('products')
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [editingGallery, setEditingGallery] = useState<GalleryItem | null>(null)
  const [editingPage, setEditingPage] = useState<Page | null>(null)
  const [showProductForm, setShowProductForm] = useState(false)
  const [showGalleryForm, setShowGalleryForm] = useState(false)
  const [showPageForm, setShowPageForm] = useState(false)
  const [uploadedImages, setUploadedImages] = useState<string[]>([])
  const [uploadingImages, setUploadingImages] = useState(false)
  const [uploadedGalleryImage, setUploadedGalleryImage] = useState<string>('')
  const [uploadingGalleryImage, setUploadingGalleryImage] = useState(false)

  useEffect(() => {
    const savedToken = localStorage.getItem('access_token')
    if (savedToken) {
      checkAdminAndSetToken(savedToken)
    }
  }, [])

  const checkAdminAndSetToken = async (t: string) => {
    try {
      // Временно устанавливаем токен для проверки
      const oldToken = localStorage.getItem('access_token')
      localStorage.setItem('access_token', t)
      
      const res = await api.isAdmin()
      if (res.isAdmin) {
        setToken(t)
      } else {
        setToken(null)
        localStorage.removeItem('access_token')
      }
    } catch (err) {
      setToken(null)
      localStorage.removeItem('access_token')
    }
  }

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    try {
      const data = await api.signIn(email, password)
      setToken(data.access)
      localStorage.setItem('access_token', data.access)

      const adminRes = await api.isAdmin()
      if (!adminRes.isAdmin) {
        throw new Error('У вас нет прав администратора')
      }
    } catch (err: any) {
      setError(err.message || 'Ошибка входа')
    } finally {
      setLoading(false)
    }
  }

  const loadProducts = async () => {
    if (!token) return
    setLoading(true)
    setError(null)
    try {
      const data = await api.adminListProducts()
      setProducts(data.items || [])
    } catch (err: any) {
      const errorMsg = err?.message || 'Ошибка загрузки товаров'
      // Если 401/403 - токен невалидный или нет прав, выходим
      if (errorMsg.includes('401') || errorMsg.includes('403') || errorMsg.includes('Unauthorized') || errorMsg.includes('Forbidden')) {
        setToken(null)
        localStorage.removeItem('access_token')
        setError('Сессия истекла. Войдите заново.')
      } else {
        setError(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  const loadGallery = async () => {
    if (!token) return
    setLoading(true)
    setError(null)
    try {
      const data = await api.adminListGalleryItems()
      setGallery(data.items || [])
    } catch (err: any) {
      const errorMsg = err?.message || 'Ошибка загрузки галереи'
      if (errorMsg.includes('401') || errorMsg.includes('403')) {
        setToken(null)
        localStorage.removeItem('access_token')
        setError('Сессия истекла. Войдите заново.')
      } else {
        setError(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  const loadPages = async () => {
    if (!token) return
    setLoading(true)
    setError(null)
    try {
      const data = await api.adminListPages()
      setPages(data.items || [])
    } catch (err: any) {
      const errorMsg = err?.message || 'Ошибка загрузки страниц'
      if (errorMsg.includes('401') || errorMsg.includes('403')) {
        setToken(null)
        localStorage.removeItem('access_token')
        setError('Сессия истекла. Войдите заново.')
      } else {
        setError(errorMsg)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (token && activeTab === 'products') loadProducts()
    if (token && activeTab === 'gallery') loadGallery()
    if (token && activeTab === 'pages') loadPages()
  }, [token, activeTab])

  const handleImageUpload = async (files: FileList | null) => {
    if (!files || files.length === 0) return

    setUploadingImages(true)
    setError(null)

    try {
      const uploadPromises = Array.from(files).map(async (file) => {
        const formData = new FormData()
        formData.append('file', file)

        const token = localStorage.getItem('access_token')
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001'}/api/admin/upload/product`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
          body: formData,
        })

        if (!response.ok) {
          const error = await response.json()
          throw new Error(error.error || 'Ошибка загрузки')
        }

        const data = await response.json()
        return data.url
      })

      const urls = await Promise.all(uploadPromises)
      setUploadedImages((prev) => [...prev, ...urls])
    } catch (err: any) {
      setError(err.message || 'Ошибка загрузки изображений')
    } finally {
      setUploadingImages(false)
    }
  }

  const handleRemoveImage = (index: number) => {
    setUploadedImages((prev) => prev.filter((_, i) => i !== index))
  }

  const handleGalleryImageUpload = async (files: FileList | null) => {
    if (!files || files.length === 0) return

    setUploadingGalleryImage(true)
    setError(null)

    try {
      const file = files[0]
      const formData = new FormData()
      formData.append('file', file)

      const token = localStorage.getItem('access_token')
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001'}/api/admin/upload/gallery`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Ошибка загрузки')
      }

      const data = await response.json()
      setUploadedGalleryImage(data.url)
    } catch (err: any) {
      setError(err.message || 'Ошибка загрузки изображения')
    } finally {
      setUploadingGalleryImage(false)
    }
  }

  const handleRemoveGalleryImage = () => {
    setUploadedGalleryImage('')
  }

  const handleCreateProduct = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const imagesStr = formData.get('images') as string
    const manualImages = imagesStr ? imagesStr.split(',').map((s) => s.trim()).filter(Boolean) : []
    const images = [...uploadedImages, ...manualImages].filter(Boolean)

    if (images.length === 0) {
      setError('Добавьте хотя бы одно изображение')
      setLoading(false)
      return
    }

    try {
      await api.adminCreateProduct({
        slug: formData.get('slug') as string,
        title: formData.get('title') as string,
        description: formData.get('description') as string,
        price: parseInt(formData.get('price') as string) * 100,
        currency: 'RUB',
        images,
        isNew: formData.get('isNew') === 'true',
        isOnSale: formData.get('isOnSale') === 'true',
      })
      setShowProductForm(false)
      setEditingProduct(null)
      setUploadedImages([])
      await loadProducts()
      // Принудительно обновляем страницу для отображения изменений
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateProduct = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!editingProduct) return

    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const imagesStr = formData.get('images') as string
    const manualImages = imagesStr ? imagesStr.split(',').map((s) => s.trim()).filter(Boolean) : []
    const images = [...uploadedImages, ...manualImages].filter(Boolean)

    try {
      await api.adminUpdateProduct(editingProduct.id, {
        slug: formData.get('slug') as string,
        title: formData.get('title') as string,
        description: formData.get('description') as string,
        price: parseInt(formData.get('price') as string) * 100,
        currency: 'RUB',
        images: images.length > 0 ? images : editingProduct.images,
        isNew: formData.get('isNew') === 'true',
        isOnSale: formData.get('isOnSale') === 'true',
      })
      setShowProductForm(false)
      setEditingProduct(null)
      setUploadedImages([])
      await loadProducts()
      // Принудительно обновляем страницу для отображения изменений
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteProduct = async (id: string) => {
    if (!confirm('Удалить товар?')) return
    setLoading(true)
    try {
      await api.adminDeleteProduct(id)
      await loadProducts()
      // Принудительно обновляем страницу для отображения изменений
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateGallery = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const imageStr = formData.get('image') as string
    const manualImage = imageStr ? imageStr.trim() : ''
    const image = uploadedGalleryImage || manualImage

    if (!image) {
      setError('Добавьте изображение')
      setLoading(false)
      return
    }

    try {
      await api.adminCreateGalleryItem({
        category: formData.get('category') as string,
        title: formData.get('title') as string,
        image: image,
        order: parseInt(formData.get('order') as string) || 0,
      })
      setShowGalleryForm(false)
      setEditingGallery(null)
      setUploadedGalleryImage('')
      loadGallery()
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateGallery = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!editingGallery) return

    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const imageStr = formData.get('image') as string
    const manualImage = imageStr ? imageStr.trim() : ''
    const image = uploadedGalleryImage || manualImage || editingGallery.image

    try {
      await api.adminUpdateGalleryItem(editingGallery.id, {
        category: formData.get('category') as string,
        title: formData.get('title') as string,
        image: image,
        order: parseInt(formData.get('order') as string) || 0,
      })
      setShowGalleryForm(false)
      setEditingGallery(null)
      setUploadedGalleryImage('')
      loadGallery()
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteGallery = async (id: string) => {
    if (!confirm('Удалить элемент галереи?')) return
    setLoading(true)
    try {
      await api.adminDeleteGalleryItem(id)
      loadGallery()
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleUpdatePage = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!editingPage) return

    setLoading(true)
    setError(null)

    const formData = new FormData(e.currentTarget)

    try {
      await api.adminUpdatePage(editingPage.slug, {
        title: formData.get('title') as string,
        content: formData.get('content') as string,
      })
      setShowPageForm(false)
      setEditingPage(null)
      loadPages()
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (!token) {
    return (
      <div className={styles.loginContainer}>
        <h1>Админ-панель</h1>
        <form onSubmit={handleLogin} className={styles.loginForm}>
          <input type="email" name="email" placeholder="Email" required />
          <input type="password" name="password" placeholder="Пароль" required />
          <button type="submit" disabled={loading}>
            {loading ? 'Вход...' : 'Войти'}
          </button>
          {error && <div className={styles.error}>{error}</div>}
        </form>
        <p className={styles.hint}>Тестовый админ: admin@socialsh.ru / admin123</p>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1>Админ-панель</h1>
        <button onClick={() => { setToken(null); localStorage.removeItem('access_token') }}>Выйти</button>
      </header>

      <nav className={styles.tabs}>
        <button className={activeTab === 'products' ? styles.active : ''} onClick={() => setActiveTab('products')}>
          Товары
        </button>
        <button className={activeTab === 'gallery' ? styles.active : ''} onClick={() => setActiveTab('gallery')}>
          Галерея
        </button>
        <button className={activeTab === 'pages' ? styles.active : ''} onClick={() => setActiveTab('pages')}>
          Страницы
        </button>
      </nav>

      {error && <div className={styles.error}>{error}</div>}

      {activeTab === 'products' && (
        <div className={styles.content}>
          <div className={styles.headerRow}>
            <h2>Товары ({products.length})</h2>
            <button onClick={() => { setEditingProduct(null); setUploadedImages([]); setError(null); setShowProductForm(true) }}>+ Создать</button>
            <button onClick={loadProducts} disabled={loading}>
              Обновить
            </button>
          </div>

          {showProductForm && (
            <div className={styles.modal}>
              <div className={styles.modalContent}>
                <h3>{editingProduct ? 'Редактировать товар' : 'Создать товар'}</h3>
                <form onSubmit={editingProduct ? handleUpdateProduct : handleCreateProduct}>
                  <input name="slug" placeholder="slug" defaultValue={editingProduct?.slug} required />
                  <input name="title" placeholder="Название" defaultValue={editingProduct?.title} required />
                  <textarea name="description" placeholder="Описание" defaultValue={editingProduct?.description} />
                  <input type="number" name="price" placeholder="Цена (руб)" defaultValue={editingProduct ? editingProduct.price / 100 : ''} required />
                  
                  <div className={styles.uploadSection}>
                    <label className={styles.uploadLabel}>
                      <input
                        type="file"
                        multiple
                        accept="image/*"
                        onChange={(e) => handleImageUpload(e.target.files)}
                        disabled={uploadingImages}
                        style={{ display: 'none' }}
                      />
                      <span className={styles.uploadButton}>
                        {uploadingImages ? 'Загрузка...' : '📷 Загрузить изображения'}
                      </span>
                    </label>
                    
                    {uploadedImages.length > 0 && (
                      <div className={styles.uploadedImages}>
                        {uploadedImages.map((url, index) => (
                          <div key={index} className={styles.uploadedImageItem}>
                            <img src={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001'}${url}`} alt={`Uploaded ${index + 1}`} />
                            <button type="button" onClick={() => handleRemoveImage(index)} className={styles.removeImageBtn}>
                              ×
                            </button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  <input 
                    name="images" 
                    placeholder="Или URL изображений (через запятую)" 
                    defaultValue={editingProduct?.images?.filter(img => !uploadedImages.includes(img)).join(', ')} 
                  />
                  
                  <label>
                    <input type="checkbox" name="isNew" value="true" defaultChecked={editingProduct?.isNew} />
                    Новинка
                  </label>
                  <label>
                    <input type="checkbox" name="isOnSale" value="true" defaultChecked={editingProduct?.isOnSale} />
                    Скидка
                  </label>
                  <div className={styles.formActions}>
                    <button type="submit" disabled={loading || uploadingImages}>
                      {loading ? 'Сохранение...' : 'Сохранить'}
                    </button>
                    <button type="button" onClick={() => { setShowProductForm(false); setEditingProduct(null); setUploadedImages([]) }}>
                      Отмена
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          <table className={styles.table}>
            <thead>
              <tr>
                <th>Slug</th>
                <th>Название</th>
                <th>Цена</th>
                <th>Новинка</th>
                <th>Скидка</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id}>
                  <td>{p.slug}</td>
                  <td>{p.title}</td>
                  <td>{p.price / 100} ₽</td>
                  <td>{p.isNew ? '✓' : ''}</td>
                  <td>{p.isOnSale ? '✓' : ''}</td>
                  <td>
                    <button onClick={() => { setEditingProduct(p); setUploadedImages([]); setError(null); setShowProductForm(true) }}>Изменить</button>
                    <button onClick={() => handleDeleteProduct(p.id)}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === 'gallery' && (
        <div className={styles.content}>
          <div className={styles.headerRow}>
            <h2>Галерея ({gallery.length})</h2>
            <button onClick={() => { setEditingGallery(null); setUploadedGalleryImage(''); setShowGalleryForm(true) }}>+ Создать</button>
            <button onClick={loadGallery} disabled={loading}>
              Обновить
            </button>
          </div>

          {showGalleryForm && (
            <div className={styles.modal}>
              <div className={styles.modalContent}>
                <h3>{editingGallery ? 'Редактировать элемент' : 'Создать элемент'}</h3>
                <form onSubmit={editingGallery ? handleUpdateGallery : handleCreateGallery}>
                  <input name="category" placeholder="Категория (intro, tattoo, tokyo...)" defaultValue={editingGallery?.category} required />
                  <input name="title" placeholder="Название" defaultValue={editingGallery?.title} />
                  
                  <div className={styles.uploadSection}>
                    <label className={styles.uploadLabel}>
                      Загрузить изображение
                      <input
                        type="file"
                        accept="image/*"
                        multiple={false}
                        onChange={(e) => handleGalleryImageUpload(e.target.files)}
                        disabled={uploadingGalleryImage}
                        style={{ display: 'none' }}
                      />
                      <button type="button" className={styles.uploadButton} disabled={uploadingGalleryImage}>
                        {uploadingGalleryImage ? 'Загрузка...' : 'Выбрать файл'}
                      </button>
                    </label>
                    
                    {uploadedGalleryImage && (
                      <div className={styles.uploadedImages}>
                        <div className={styles.uploadedImageItem}>
                          <img src={uploadedGalleryImage} alt="Preview" style={{ width: '100px', height: '100px', objectFit: 'contain' }} />
                          <button type="button" onClick={handleRemoveGalleryImage} className={styles.removeImageBtn}>×</button>
                        </div>
                      </div>
                    )}
                  </div>

                  <input 
                    name="image" 
                    placeholder="Или введите URL изображения" 
                    defaultValue={editingGallery?.image && !uploadedGalleryImage ? editingGallery.image : ''} 
                  />
                  <input type="number" name="order" placeholder="Порядок" defaultValue={editingGallery?.order || 0} />
                  <div className={styles.formActions}>
                    <button type="submit" disabled={loading || uploadingGalleryImage}>
                      {loading ? 'Сохранение...' : 'Сохранить'}
                    </button>
                    <button type="button" onClick={() => { setShowGalleryForm(false); setEditingGallery(null); setUploadedGalleryImage('') }}>
                      Отмена
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          <div className={styles.galleryGrid}>
            {gallery.map((item) => (
              <div key={item.id} className={styles.galleryItem}>
                {item.image && <img src={item.image} alt={item.title} className={styles.galleryImage} />}
                <div className={styles.galleryInfo}>
                  <div>{item.title || 'Без названия'}</div>
                  <div className={styles.category}>{item.category}</div>
                  <div className={styles.galleryActions}>
                    <button onClick={() => { setEditingGallery(item); setUploadedGalleryImage(''); setShowGalleryForm(true) }}>Изменить</button>
                    <button onClick={() => handleDeleteGallery(item.id)}>Удалить</button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'pages' && (
        <div className={styles.content}>
          <h2>Страницы ({pages.length})</h2>
          <button onClick={loadPages} disabled={loading}>
            Обновить
          </button>

          {showPageForm && editingPage && (
            <div className={styles.modal}>
              <div className={styles.modalContent}>
                <h3>Редактировать страницу: {editingPage.slug}</h3>
                <form onSubmit={handleUpdatePage}>
                  <input name="title" placeholder="Заголовок" defaultValue={editingPage.title} required />
                  <textarea name="content" placeholder="Содержимое (HTML)" defaultValue={editingPage.content} rows={10} required />
                  <div className={styles.formActions}>
                    <button type="submit" disabled={loading}>
                      {loading ? 'Сохранение...' : 'Сохранить'}
                    </button>
                    <button type="button" onClick={() => { setShowPageForm(false); setEditingPage(null) }}>
                      Отмена
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          <table className={styles.table}>
            <thead>
              <tr>
                <th>Slug</th>
                <th>Заголовок</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              {pages.map((page) => (
                <tr key={page.slug}>
                  <td>{page.slug}</td>
                  <td>{page.title}</td>
                  <td>
                    <button onClick={() => { setEditingPage(page); setShowPageForm(true) }}>Редактировать</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
