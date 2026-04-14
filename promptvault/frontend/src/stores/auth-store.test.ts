import { useAuthStore } from "./auth-store"
import { clearTokens, getAccessToken } from "@/api/client"
import type { AuthResponse, User } from "@/api/types"

// Mock the entire api/client module
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>(
    "@/api/client",
  )
  return {
    ...actual,
    api: vi.fn(),
    apiVoid: vi.fn(),
    ensureFreshToken: vi.fn(),
  }
})

// Import mocked functions after vi.mock
import { api, apiVoid, ensureFreshToken } from "@/api/client"

const mockedApi = vi.mocked(api)
const mockedApiVoid = vi.mocked(apiVoid)
const mockedEnsureFreshToken = vi.mocked(ensureFreshToken)

const testUser: User = {
  id: 1,
  email: "test@example.com",
  name: "Test User",
  email_verified: true,
  has_password: true,
  default_model: "anthropic/claude-sonnet-4",
}

const testAuthResponse: AuthResponse = {
  user: testUser,
  tokens: {
    access_token: "test-access",
    expires_in: 900,
  },
}

beforeEach(() => {
  vi.clearAllMocks()
  clearTokens()
  // Reset store to initial state
  useAuthStore.setState({ user: null, isLoading: true })
})

describe("auth-store", () => {
  describe("login", () => {
    it("sets user and tokens on success", async () => {
      mockedApi.mockResolvedValueOnce(testAuthResponse)
      mockedApi.mockResolvedValueOnce(testUser) // fetchMe after login

      await useAuthStore.getState().login("test@example.com", "password123")

      expect(mockedApi).toHaveBeenCalledWith("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email: "test@example.com", password: "password123" }),
      })

      const state = useAuthStore.getState()
      expect(state.user).toEqual(testUser)
    })
  })

  describe("register", () => {
    it("does not set user or tokens", async () => {
      mockedApi.mockResolvedValueOnce({
        email: "test@example.com",
        message: "verification email sent",
      })

      await useAuthStore.getState().register("test@example.com", "password123", "Test")

      expect(mockedApi).toHaveBeenCalledWith("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email: "test@example.com", password: "password123", name: "Test" }),
      })

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
    })
  })

  describe("logout", () => {
    it("clears user state", async () => {
      // Set initial authenticated state
      useAuthStore.setState({ user: testUser })
      expect(useAuthStore.getState().user).toEqual(testUser)

      mockedApiVoid.mockResolvedValueOnce(undefined)

      await useAuthStore.getState().logout()

      expect(useAuthStore.getState().user).toBeNull()
      expect(getAccessToken()).toBeNull()
    })

    it("clears user even if logout API call fails", async () => {
      useAuthStore.setState({ user: testUser })
      mockedApiVoid.mockRejectedValueOnce(new Error("network error"))

      await useAuthStore.getState().logout()

      expect(useAuthStore.getState().user).toBeNull()
    })
  })

  describe("fetchMe", () => {
    it("sets user on success", async () => {
      mockedApi.mockResolvedValueOnce(testUser)

      await useAuthStore.getState().fetchMe()

      expect(mockedApi).toHaveBeenCalledWith("/auth/me")
      expect(useAuthStore.getState().user).toEqual(testUser)
    })
  })

  describe("isAuthenticated", () => {
    it("returns true when user is set", () => {
      useAuthStore.setState({ user: testUser, isAuthenticated: true })
      expect(useAuthStore.getState().isAuthenticated).toBe(true)
    })

    it("returns false when user is null", () => {
      useAuthStore.setState({ user: null, isAuthenticated: false })
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })
  })

  describe("restoreSession", () => {
    it("success — sets user and stops loading", async () => {
      localStorage.setItem("pv_has_session", "1")
      mockedEnsureFreshToken.mockResolvedValueOnce(undefined)
      mockedApi.mockResolvedValueOnce(testUser)

      await useAuthStore.getState().restoreSession()

      expect(mockedEnsureFreshToken).toHaveBeenCalledTimes(1)
      expect(mockedApi).toHaveBeenCalledWith("/auth/me")

      const state = useAuthStore.getState()
      expect(state.user).toEqual(testUser)
      expect(state.isLoading).toBe(false)
    })

    it("auth error clears state and sets sessionError=auth", async () => {
      localStorage.setItem("pv_has_session", "1")
      mockedEnsureFreshToken.mockRejectedValueOnce(new Error("Сессия истекла"))

      await useAuthStore.getState().restoreSession()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.isLoading).toBe(false)
      expect(state.sessionError).toBe("auth")
    })

    it("unauthorized error clears state", async () => {
      localStorage.setItem("pv_has_session", "1")
      mockedEnsureFreshToken.mockRejectedValueOnce(new Error("unauthorized"))

      await useAuthStore.getState().restoreSession()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isLoading).toBe(false)
      expect(state.sessionError).toBe("auth")
    })

    it("transient error retries 3 times and sets sessionError=transient", async () => {
      localStorage.setItem("pv_has_session", "1")
      useAuthStore.setState({ user: testUser, isLoading: true })
      // Все 3 попытки падают с transient ошибкой — user должен остаться нетронутым.
      mockedEnsureFreshToken
        .mockRejectedValueOnce(new Error("transient: network unavailable"))
        .mockRejectedValueOnce(new Error("transient: server 500"))
        .mockRejectedValueOnce(new Error("transient: network unavailable"))

      await useAuthStore.getState().restoreSession()

      expect(mockedEnsureFreshToken).toHaveBeenCalledTimes(3)
      const state = useAuthStore.getState()
      expect(state.user).toEqual(testUser)
      expect(state.isLoading).toBe(false)
      expect(state.sessionError).toBe("transient")
    })
  })
})
