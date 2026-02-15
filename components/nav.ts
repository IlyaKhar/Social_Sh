export type NavGroup = {
  title: string
  items: Array<{ label: string; href: string }>
}

export const NAV: NavGroup[] = [
  {
    title: 'Магазин',
    items: [
      { label: 'Новые поступления', href: '/shop/new' },
      { label: 'Сезонные скидки', href: '/shop/sale' },
      { label: 'Смотреть все', href: '/shop' }
    ]
  },
  {
    title: 'Информация',
    items: [
      { label: 'Оплата', href: '/info/payment' },
      { label: 'Доставка', href: '/info/delivery' },
      { label: 'Вакансии', href: '/info/jobs' },
      { label: 'Контакты', href: '/contacts' },
      { label: 'Возврат', href: '/info/returns' }
    ]
  },
  {
    title: 'Галерея',
    items: [{ label: 'Смотреть', href: '/gallery' }]
  },
  {
    title: 'Проекты',
    items: [{ label: 'Смотреть', href: '/projects' }]
  }
]

