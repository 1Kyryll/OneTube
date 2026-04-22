import { AuthProvider } from "@/features/auth/auth-provider";

export const metadata = {
  title: "OneTube",
  description: "Demo video app",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
