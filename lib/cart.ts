// Утилиты для работы с корзиной (localStorage)

export interface CartItem {
  productId: string
  slug: string
  title: string
  price: number
  currency: string
  image: string
  quantity: number
}

const CART_KEY = 'socialsh_cart'

// Получить корзину из localStorage
export function getCart(): CartItem[] {
  if (typeof window === 'undefined') return []
  try {
    const cart = localStorage.getItem(CART_KEY)
    return cart ? JSON.parse(cart) : []
  } catch {
    return []
  }
}

// Сохранить корзину в localStorage
export function saveCart(cart: CartItem[]): void {
  if (typeof window === 'undefined') return
  try {
    localStorage.setItem(CART_KEY, JSON.stringify(cart))
  } catch (err) {
    console.error('Failed to save cart:', err)
  }
}

// Добавить товар в корзину
export function addToCart(item: Omit<CartItem, 'quantity'>): void {
  const cart = getCart()
  const existing = cart.find((i) => i.productId === item.productId)

  if (existing) {
    existing.quantity += 1
  } else {
    cart.push({ ...item, quantity: 1 })
  }

  saveCart(cart)
}

// Удалить товар из корзины
export function removeFromCart(productId: string): void {
  const cart = getCart().filter((i) => i.productId !== productId)
  saveCart(cart)
}

// Изменить количество товара
export function updateCartItemQuantity(productId: string, quantity: number): void {
  if (quantity <= 0) {
    removeFromCart(productId)
    return
  }

  const cart = getCart()
  const item = cart.find((i) => i.productId === productId)
  if (item) {
    item.quantity = quantity
    saveCart(cart)
  }
}

// Очистить корзину
export function clearCart(): void {
  saveCart([])
}

// Получить общее количество товаров в корзине
export function getCartTotalItems(): number {
  return getCart().reduce((sum, item) => sum + item.quantity, 0)
}

// Получить общую сумму корзины
export function getCartTotalPrice(): number {
  return getCart().reduce((sum, item) => sum + item.price * item.quantity, 0)
}
