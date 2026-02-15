import type { ReactNode } from 'react'
import styles from './Container.module.css'

export function Container(props: { children: ReactNode; size?: 'default' | 'wide' }) {
  const { children, size = 'default' } = props

  return <div className={size === 'wide' ? styles.wide : styles.root}>{children}</div>
}

