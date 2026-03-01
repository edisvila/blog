import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50">
      <h1 className="text-4xl font-bold text-gray-900 mb-6">
        Hello World mit Tailwind!
      </h1>

      <button
        className="px-6 py-3 bg-gray-500 text-white font-semibold rounded-lg shadow hover:bg-blue-600 transition"
        onClick={() => setCount(count + 1)}
      >
        Du hast {count} mal geklickt
      </button>

      <p className="mt-4 text-gray-700">
        Edit <code className="bg-gray-200 px-1 rounded">src/App.tsx</code> und
        speichere, um HMR zu testen
      </p>
    </div>
  )
}

export default App
