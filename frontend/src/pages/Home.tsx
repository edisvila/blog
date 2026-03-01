import { Link } from 'react-router-dom'
import { posts } from '../posts'

export default function Home() {
  return (
    <main className="max-w-2xl mx-auto px-4 py-16">
      <h1 className="text-2xl font-bold mb-2">edisvila.de</h1>
      <p className="text-gray-500 mb-12">Notizen aus der Entwicklung</p>

      <ul className="space-y-8">
        {posts.map(post => (
          <li key={post.slug}>
            <span className="text-xs text-gray-400 uppercase tracking-wide">
              {post.category}
            </span>
            <Link
              to={`/posts/${post.slug}`}
              className="block mt-1 text-lg hover:underline"
            >
              {post.title}
            </Link>
          </li>
        ))}
      </ul>
    </main>
  )
}