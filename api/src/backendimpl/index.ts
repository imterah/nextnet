import { BackendBaseClass } from "./base.js";

import { PassyFireBackendProvider } from "./passyfire-reimpl/index.js";
import { SSHPyBackendProvider } from "./sshpy.js";
import { SSHBackendProvider } from "./ssh.js";

export const backendProviders: Record<string, typeof BackendBaseClass> = {
  ssh: SSHBackendProvider,
  sshpy: SSHPyBackendProvider,
  passyfire: PassyFireBackendProvider,
};

if (process.env.NODE_ENV != "production") {
  backendProviders["dummy"] = BackendBaseClass;
}
