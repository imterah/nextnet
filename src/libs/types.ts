export type ServerOptions = {
  isSignupEnabled: boolean;
  isSignupAsAdminEnabled: boolean;

  allowUnsafeGlobalTokens: boolean;
}

// NOTE: Someone should probably use Redis for this, but this is fine...
export type SessionToken = {
  createdAt: number,
  expiresAt: number, // Should be (createdAt + (30 minutes))
  
  token: string
};