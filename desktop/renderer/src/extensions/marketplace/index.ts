import { extensionRegistry } from '../registry'
import { apiStyleSpecExtension } from './api-style-spec'

export function registerMarketplaceExtensions() {
  extensionRegistry.register(apiStyleSpecExtension, 'marketplace')
}

export { MarketplaceClient, marketplaceClient } from './client'
