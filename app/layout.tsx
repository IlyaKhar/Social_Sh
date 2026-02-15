import type { Metadata } from 'next'
import { IBM_Plex_Mono, Space_Grotesk } from 'next/font/google'
import './globals.css'
import { PageHeader } from '@/components/PageHeader'

const sans = Space_Grotesk({
  subsets: ['latin', 'latin-ext'],
  weight: ['400', '500', '600'],
  variable: '--font-sans'
})

const mono = IBM_Plex_Mono({
  subsets: ['latin', 'latin-ext'],
  weight: ['400', '500'],
  variable: '--font-mono'
})

export const metadata: Metadata = {
  title: 'SOCIAL SH',
  description: 'Store redesign'
}

export default function RootLayout(props: { children: React.ReactNode }) {
  const { children } = props

  return (
    <html lang="ru" className={`${sans.variable} ${mono.variable}`}>
      <body>
        <a className="skipLink" href="#main">
          Пропустить к содержимому
        </a>
        <div className="shell">
          <main id="main" className="main">
            {children}
          </main>
          <aside className="sideHeader" aria-label="Навигация сайта">
            <PageHeader />
          </aside>
        </div>
      </body>
    </html>
  )
}

