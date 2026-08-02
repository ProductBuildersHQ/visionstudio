import { extensionRegistry } from '../registry'
import { visionspecExtension } from './visionspec'
import { aidlcExtension } from './aidlc'
import { v2momExtension } from './v2mom'
import { analyticsExtension } from './analytics'
import { unifiedDashboardExtension } from './unified-dashboard'

export function registerBuiltinExtensions() {
  extensionRegistry.register(unifiedDashboardExtension, 'builtin')
  extensionRegistry.register(visionspecExtension, 'builtin')
  extensionRegistry.register(aidlcExtension, 'builtin')
  extensionRegistry.register(v2momExtension, 'builtin')
  extensionRegistry.register(analyticsExtension, 'builtin')
}
