import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Loader2, Check } from "lucide-react"
import { toast } from "sonner"

import { useAuthStore } from "@/stores/auth-store"
import { useUpdateProfile } from "@/hooks/use-settings"
import { Button } from "@/components/ui/button"
import { SectionHeader } from "./_section-header"

const profileSchema = z.object({
  name: z.string().min(1, "Введите имя").max(100),
  username: z
    .string()
    .max(30)
    .regex(/^[a-zA-Z0-9_]*$/, "Только латинские буквы, цифры и _")
    .optional()
    .or(z.literal("")),
})

export default function SettingsProfilePage() {
  const user = useAuthStore((s) => s.user)
  const fetchMe = useAuthStore((s) => s.fetchMe)
  const updateProfile = useUpdateProfile()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user?.name ?? "", username: user?.username ?? "" },
  })

  if (!user) return null

  const onSubmit = async (data: { name: string; username?: string }) => {
    try {
      await updateProfile.mutateAsync({ name: data.name, username: data.username || undefined })
      fetchMe()
      toast.success("Профиль обновлён")
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Ошибка обновления")
    }
  }

  const initials =
    user.name
      .split(" ")
      .filter(Boolean)
      .map((w) => w[0])
      .join("")
      .toUpperCase()
      .slice(0, 2) || "?"

  return (
    <section>
      <SectionHeader title="Профиль" description="Имя, никнейм и аватар" />

      <div className="mb-6 flex items-center gap-4">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt={user.name} className="h-14 w-14 rounded-full object-cover" />
        ) : (
          <div className="flex h-14 w-14 items-center justify-center rounded-full text-lg font-semibold text-brand-foreground [background:var(--brand-gradient)]">
            {initials}
          </div>
        )}
        <div>
          <p className="text-sm font-medium text-foreground">{user.name}</p>
          <p className="text-xs text-muted-foreground">{user.email}</p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-3 max-w-md">
        <div>
          <label className="text-[0.75rem] text-muted-foreground" htmlFor="profile-name">Имя</label>
          <input
            id="profile-name"
            {...register("name")}
            className="mt-1 h-11 w-full rounded-lg border border-border bg-background px-3 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
          />
          {errors.name && <p className="mt-1 text-xs text-red-400">{errors.name.message}</p>}
        </div>

        <div>
          <label className="text-[0.75rem] text-muted-foreground" htmlFor="profile-username">Никнейм</label>
          <div className="relative mt-1">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">@</span>
            <input
              id="profile-username"
              {...register("username")}
              placeholder="username"
              className="h-11 w-full rounded-lg border border-border bg-background pl-7 pr-3 text-sm text-foreground outline-none transition-colors focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
            />
          </div>
          {errors.username && <p className="mt-1 text-xs text-red-400">{errors.username.message}</p>}
        </div>

        <div>
          <label className="text-[0.75rem] text-muted-foreground" htmlFor="profile-email">Email</label>
          <input
            id="profile-email"
            value={user.email}
            disabled
            className="mt-1 h-11 w-full rounded-lg border border-border bg-muted px-3 text-sm text-muted-foreground outline-none cursor-not-allowed"
          />
        </div>

        <Button type="submit" variant="brand" disabled={updateProfile.isPending}>
          {updateProfile.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
          Сохранить
        </Button>
      </form>
    </section>
  )
}
