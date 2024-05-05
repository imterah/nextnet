import type { BackendBaseClass } from "./base.js";

import { PassyFireBackendProvider } from "./passyfire-reimpl/index.js";
import { SSHBackendProvider } from "./ssh.js";

export const backendProviders: Record<string, typeof BackendBaseClass> = {
  "ssh": SSHBackendProvider,
  "passyfire": PassyFireBackendProvider
};