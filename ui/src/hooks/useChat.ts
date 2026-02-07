import { useSearchParams } from 'react-router-dom'

export function useChat() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTalker = searchParams.get('talker')

  const setActiveTalker = (talker: string) => {
    setSearchParams({ talker })
  }

  return {
    activeTalker,
    setActiveTalker
  }
}
