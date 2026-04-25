import { AuthProvider } from "@/features/auth/auth-provider";
import { NavBar } from "./_components/NavBar";

export const metadata = {
  title: "OneTube",
  description: "Demo video app",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body style={{ margin: 0, background: "#0a0a0a", color: "#fff", fontFamily: "system-ui, sans-serif" }}>
        <AuthProvider>
          <NavBar />
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
