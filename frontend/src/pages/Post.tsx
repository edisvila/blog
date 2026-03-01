import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { posts } from '../posts'
import { renderMarkdown } from '../lib/markdown'

export default function Post() {
  const { slug } = useParams()
  const [html, setHtml] = useState('')
  const meta = posts.find(p => p.slug === slug)

  useEffect(() => {
    if (!meta) return
    renderMarkdown(meta.content).then(it => setHtml(it))
  }, [meta])

  if (!meta) return <div>Post nicht gefunden</div>

  return (
    <main className="max-w-2xl mx-auto px-4 py-16">
      <Link to="/" className="text-sm text-gray-400 hover:underline mb-8 block">
        ← Zurück
      </Link>

      <article
        className="prose prose-gray max-w-none"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </main>
  )
}