const KEY = 'astrophage.token'
const EMAIL_KEY = 'astrophage.email'

// localStorage can throw (private mode, blocked storage): the app then just asks for login again.
export function getToken(): string | null {
  try {
    return localStorage.getItem(KEY)
  } catch {
    return null
  }
}

export function getEmail(): string {
  try {
    return localStorage.getItem(EMAIL_KEY) ?? ''
  } catch {
    return ''
  }
}

export function saveSession(token: string, email: string) {
  try {
    localStorage.setItem(KEY, token)
    localStorage.setItem(EMAIL_KEY, email)
  } catch {
    /* ignore */
  }
}

export function clearSession() {
  try {
    localStorage.removeItem(KEY)
    localStorage.removeItem(EMAIL_KEY)
  } catch {
    /* ignore */
  }
}
