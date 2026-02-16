// Утилиты для работы с API

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001'

// Формирует полный URL для изображения (если путь относительный, добавляет базовый URL API)
export function getImageUrl(imagePath: string | undefined | null): string {
  if (!imagePath) return getPlaceholderImage(400, 500)
  
  // Если уже полный URL (http:// или https://), возвращаем как есть
  if (imagePath.startsWith('http://') || imagePath.startsWith('https://')) {
    return imagePath
  }
  
  // Если путь начинается с /uploads, добавляем базовый URL API
  if (imagePath.startsWith('/uploads')) {
    return `${API_URL}${imagePath}`
  }
  
  // Иначе возвращаем как есть (может быть placeholder)
  return imagePath
}

// Типы для API ответов
export type Product = {
  id: string
  slug: string
  title: string
  description: string
  price: number
  currency: string
  images: string[]
  isNew: boolean
  isOnSale: boolean
}

export type GalleryItem = {
  id: string
  category: string
  title: string
  image: string
  order: number
}

export type Page = {
  slug: string
  title: string
  content: string
}

export type User = {
  id: string
  email: string
  name: string
  role: string
}

export type Order = {
  id: string
  userId: string
  status: string
  total: number
  createdAt: string
  items: OrderItem[]
}

export type OrderItem = {
  id: string
  productId: string
  title: string
  price: number
  quantity: number
}

export type ProductsResponse = {
  items: Product[]
}

export type GalleryResponse = {
  items: GalleryItem[]
}

export type PagesResponse = {
  items: Page[]
}

export type OrdersResponse = {
  items: Order[]
}

// Получить токен из localStorage
function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('access_token')
}

// Базовый fetch с обработкой ошибок
async function fetchAPI<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  // Отключаем кеширование для всех запросов - всегда получаем свежие данные
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
    cache: 'no-store', // Не кешируем запросы
  })

  if (!response.ok) {
    let errorMessage = `HTTP ${response.status}`
    try {
      const error = await response.json()
      errorMessage = error.error || error.message || errorMessage
    } catch {
      errorMessage = response.statusText || errorMessage
    }
    
    // Если 401 или 403 - удаляем токен (он невалидный или недостаточно прав)
    if (response.status === 401 || response.status === 403) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token')
      }
    }
    
    throw new Error(errorMessage)
  }

  return response.json()
}

// Публичные эндпоинты
export const api = {
  // Товары
  getProducts: (params?: { new?: boolean; sale?: boolean; page?: number; limit?: number }) => {
    const query = new URLSearchParams()
    if (params?.new) query.append('new', 'true')
    if (params?.sale) query.append('sale', 'true')
    if (params?.page) query.append('page', params.page.toString())
    if (params?.limit) query.append('limit', params.limit.toString())
    return fetchAPI<{ items: Product[] }>(`/api/products?${query.toString()}`)
  },

  getProduct: (slug: string) => {
    return fetchAPI<{ item: Product }>(`/api/products/${slug}`)
  },

  searchProducts: (query: string, page: number = 1, limit: number = 20) => {
    return fetchAPI<{ items: Product[] }>(
      `/api/products/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`
    )
  },

  // Галерея
  getGalleryItems: (category?: string) => {
    const query = category ? `?category=${category}` : ''
    return fetchAPI<{ items: GalleryItem[] }>(`/api/gallery${query}`)
  },

  // Страницы
  getPage: (slug: string) => {
    return fetchAPI<Page>(`/api/pages/${slug}`)
  },

  // Авторизация
  signIn: (email: string, password: string) => {
    return fetchAPI<{ access: string; refresh: string }>('/api/auth/sign-in', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },

  signUp: (email: string, password: string, name: string) => {
    return fetchAPI<{ access: string; refresh: string }>('/api/auth/sign-up', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    })
  },

  isAdmin: async () => {
    try {
      return await fetchAPI<{ isAdmin: boolean }>('/api/auth/is-admin')
    } catch (err) {
      // Если 401 - токен невалидный, возвращаем false
      if (err instanceof Error && err.message.includes('401')) {
        return { isAdmin: false }
      }
      throw err
    }
  },

  // Личный кабинет
  getAccount: () => {
    return fetchAPI<{ user: User }>('/api/account/me')
  },

  getOrders: () => {
    return fetchAPI<{ items: Order[] }>('/api/account/orders')
  },

  updateProfile: (data: { name?: string; email?: string }) => {
    return fetchAPI<{ user: User }>('/api/account/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  // Админка - Товары
  adminListProducts: () => {
    return fetchAPI<{ items: Product[] }>('/api/admin/products')
  },

  adminGetProduct: (id: string) => {
    return fetchAPI<Product>(`/api/admin/products/${id}`)
  },

  adminCreateProduct: (product: Omit<Product, 'id'>) => {
    return fetchAPI<Product>('/api/admin/products', {
      method: 'POST',
      body: JSON.stringify(product),
    })
  },

  adminUpdateProduct: (id: string, product: Partial<Product>) => {
    return fetchAPI<Product>(`/api/admin/products/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(product),
    })
  },

  adminDeleteProduct: (id: string) => {
    return fetchAPI<{ message: string }>(`/api/admin/products/${id}`, {
      method: 'DELETE',
    })
  },

  // Админка - Галерея
  adminListGalleryItems: () => {
    return fetchAPI<{ items: GalleryItem[] }>('/api/admin/gallery')
  },

  adminCreateGalleryItem: (item: Omit<GalleryItem, 'id'>) => {
    return fetchAPI<GalleryItem>('/api/admin/gallery', {
      method: 'POST',
      body: JSON.stringify(item),
    })
  },

  adminUpdateGalleryItem: (id: string, item: Partial<GalleryItem>) => {
    return fetchAPI<GalleryItem>(`/api/admin/gallery/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(item),
    })
  },

  adminDeleteGalleryItem: (id: string) => {
    return fetchAPI<{ message: string }>(`/api/admin/gallery/${id}`, {
      method: 'DELETE',
    })
  },

  // Админка - Страницы
  adminListPages: () => {
    return fetchAPI<{ items: Page[] }>('/api/admin/pages')
  },

  adminUpdatePage: (slug: string, page: Partial<Page>) => {
    return fetchAPI<Page>(`/api/admin/pages/${slug}`, {
      method: 'PATCH',
      body: JSON.stringify(page),
    })
  },

  // Создание заказа
  createOrder: (order: {
    items: Array<{ productId: string; quantity: number; price: number }>
    customer: {
      name: string
      email: string
      phone?: string
      telegram?: string
      address?: string
    }
    comment?: string
    total: number
  }) => {
    return fetchAPI<{ message: string; orderId?: string }>('/api/orders', {
      method: 'POST',
      body: JSON.stringify(order),
    })
  },
}

// Форматирование цены
export function formatPrice(price: number, currency: string = 'RUB'): string {
  const formatter = new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
  })
  return formatter.format(price / 100) // цена в копейках
}

// Placeholder изображения
export function getPlaceholderImage(width: number = 400, height: number = 400): string {
  return `https://picsum.photos/${width}/${height}?random=${Math.random()}`
}
