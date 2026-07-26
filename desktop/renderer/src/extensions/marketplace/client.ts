import type { MarketplaceManifest, MarketplaceEntry } from '../../types/extension'

const DEFAULT_MANIFEST_URL =
  'https://raw.githubusercontent.com/ProductBuildersHQ/visionstudio-extensions/main/extensions.json'

export class MarketplaceClient {
  private manifestUrl: string
  private cache: MarketplaceManifest | null = null
  private cacheTime = 0
  private cacheTTL = 5 * 60 * 1000

  constructor(manifestUrl = DEFAULT_MANIFEST_URL) {
    this.manifestUrl = manifestUrl
  }

  async fetchManifest(): Promise<MarketplaceManifest> {
    const now = Date.now()
    if (this.cache && now - this.cacheTime < this.cacheTTL) {
      return this.cache
    }

    const response = await fetch(this.manifestUrl)
    if (!response.ok) {
      throw new Error(`Failed to fetch marketplace: ${response.status}`)
    }

    const manifest: MarketplaceManifest = await response.json()
    this.cache = manifest
    this.cacheTime = now
    return manifest
  }

  async listExtensions(): Promise<MarketplaceEntry[]> {
    const manifest = await this.fetchManifest()
    return manifest.extensions
  }

  async search(query: string): Promise<MarketplaceEntry[]> {
    const extensions = await this.listExtensions()
    const q = query.toLowerCase()
    return extensions.filter(
      e =>
        e.name.toLowerCase().includes(q) ||
        e.description.toLowerCase().includes(q) ||
        e.tags?.some(t => t.toLowerCase().includes(q))
    )
  }
}

export const marketplaceClient = new MarketplaceClient()
