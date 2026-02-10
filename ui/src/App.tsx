import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAppStore } from '@/stores/app'
import { MainLayout } from '@/components/layout/MainLayout'
import Chat from '@/views/Chat'
import Contact from '@/views/Contact'
import Search from '@/views/Search'
import Dashboard from '@/views/Dashboard'
import AnnualReport from '@/views/AnnualReport'
import Sentiment from '@/views/Sentiment'
import WordCloud from '@/views/WordCloud'
import { PaymentModal } from '@/components/PaymentModal'
import { AgreementModal } from '@/components/AgreementModal'

const AGREEMENT_KEY = 'wetrace_agreement_accepted'

function App() {
  const theme = useAppStore((state) => state.settings.theme)
  const setMobile = useAppStore((state) => state.setMobile)
  const [agreed, setAgreed] = useState<boolean | null>(null)

  useEffect(() => {
    // Check agreement status
    const isAgreed = localStorage.getItem(AGREEMENT_KEY) === 'true'
    setAgreed(isAgreed)

    // Initialize theme
    const isDark = theme === 'dark' || (theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
    
    if (isDark) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }

    // Initialize mobile check
    const checkMobile = () => {
      setMobile(window.innerWidth <= 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [theme, setMobile])

  const handleAcceptAgreement = () => {
    localStorage.setItem(AGREEMENT_KEY, 'true')
    setAgreed(true)
  }

  if (agreed === null) return null

  return (
    <BrowserRouter>
      {!agreed && <AgreementModal onAccept={handleAcceptAgreement} />}
      <PaymentModal />
      <Routes>
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="/chat" replace />} />
          <Route path="chat" element={<Chat />} />
          <Route path="contact" element={<Contact />} />
          <Route path="search" element={<Search />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="report" element={<AnnualReport />} />
          <Route path="sentiment" element={<Sentiment />} />
          <Route path="wordcloud" element={<WordCloud />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App