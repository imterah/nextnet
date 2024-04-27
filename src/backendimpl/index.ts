import type { BackendBaseClass } from "./base.js";
import { SSHBackendProvider } from "./ssh.js";

export const backendProviders: Record<string, typeof BackendBaseClass> = {
  "ssh": SSHBackendProvider
};