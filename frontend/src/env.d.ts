interface ImportMetaEnv {
  /**
   * Base URL of the Go backend, such as http://localhost:8080. Leave it unset
   * to use same-origin requests (the Vite proxy in dev, the Go server in Docker).
   */
  readonly VITE_API_BASE_URL?: string
}
