import { Outlet } from "react-router-dom"
import { Sidebar } from "./Sidebar"
import { MobileNav } from "./MobileNav"
import { useAppStore } from "@/stores/app"

export function MainLayout() {
  const isMobile = useAppStore((state) => state.isMobile)

  return (
    <div className="flex h-screen w-full bg-background overflow-hidden text-foreground">
      {!isMobile && <Sidebar />}
      <main className="flex-1 h-full overflow-hidden flex flex-col relative">
        <Outlet />
      </main>
      {isMobile && <MobileNav />}
    </div>
  )
}
