import blogEinfuehrung from '../posts/blog-einfuehrung.md?raw'
import serverHardening from '../posts/server-hardening.md?raw'
import nginxCertbot from '../posts/nginx-certbot-cicd.md?raw'
import openapiFirst from '../posts/openapi-first.md?raw'
import goJavaJs from '../posts/go-fuer-java-js-entwickler.md?raw'

export interface PostMeta {
  slug: string
  title: string
  category: string
  content: string  // statt 'file'
}

export const posts: PostMeta[] = [
  {
    slug: 'blog-einfuehrung',
    title: 'Warum ich meinen Blog selbst gebaut habe',
    category: 'Entwicklung',
    content: blogEinfuehrung,
  },
  {
    slug: 'server-hardening',
    title: 'Server Hardening: nftables, Docker und fail2ban',
    category: 'Infrastruktur',
    content: serverHardening,
  },
  {
    slug: 'nginx-certbot-cicd',
    title: 'TLS mit nginx und certbot',
    category: 'DevOps',
    content: nginxCertbot,
  },
  {
    slug: 'openapi-first',
    title: 'OpenAPI-First: Frontend und Backend aus einer Spec generieren',
    category: 'Entwicklung',
    content: openapiFirst,
  },
  {
    slug: 'go-fuer-java-js-entwickler',
    title: 'Go aus der Perspektive eines Java- und JavaScript-Entwicklers',
    category: 'Entwicklung',
    content: goJavaJs,
  },
]