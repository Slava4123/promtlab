import { Link } from "react-router-dom"
import { Card, CardContent } from "@/components/ui/card"
import { buttonVariants } from "@/components/ui/button"
import { Sparkles } from "lucide-react"

interface UpgradeGateProps {
  title: string
  description: string
  targetPlan: "Pro" | "Max"
}

export function UpgradeGate({ title, description, targetPlan }: UpgradeGateProps) {
  return (
    <Card className="border-dashed">
      <CardContent className="flex flex-col items-center gap-3 py-8 text-center">
        <div className="flex size-10 items-center justify-center rounded-full bg-primary/10">
          <Sparkles className="size-5 text-primary" />
        </div>
        <div>
          <h3 className="text-base font-semibold">{title}</h3>
          <p className="mt-1 max-w-md text-sm text-muted-foreground">{description}</p>
        </div>
        <Link to="/pricing" className={buttonVariants()}>
          Перейти на {targetPlan}
        </Link>
      </CardContent>
    </Card>
  )
}
