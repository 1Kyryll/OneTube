import { apiRequest } from "@/shared/api/client";
import type { LoginInput, SignupInput, User } from "./types";

export const authApi = {
  signup: (data: SignupInput) =>
    apiRequest<User>("/api/auth/signup", { method: "POST", body: JSON.stringify(data) }),
  login: (data: LoginInput) =>
    apiRequest<User>("/api/auth/login", { method: "POST", body: JSON.stringify(data) }),
  logout: () => apiRequest<void>("/api/auth/logout", { method: "POST" }),
  me: () => apiRequest<User>("/api/auth/me"),
};
