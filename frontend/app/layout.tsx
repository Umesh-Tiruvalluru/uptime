import "./globals.css"
import { ThemeProvider } from "@/components/theme-provider"
import { AuthProvider } from "@/context/auth-context"
import { Navbar } from "@/components/navbar"

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className="antialiased font-sans"
    >
      <body className="min-h-screen bg-background font-sans text-foreground selection:bg-primary/20">
        <ThemeProvider>
          <AuthProvider>
            <div className="relative flex min-h-screen flex-col">
              <Navbar />
              <div className="flex-1">{children}</div>
              <footer className="border-t border-border/40 py-6 text-center text-xs text-muted-foreground">
                <div className="mx-auto max-w-6xl px-4 flex items-center justify-center">
                  <span>Uptime Monitor © {new Date().getFullYear()}</span>
                </div>
              </footer>
            </div>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
