import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, Save, Lock } from "lucide-react"
import { useMutation } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { useAuthStore } from "../../stores/auth-store"
import type { UpdateProfileBody } from "../../lib/api"

export function ProfilePage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const [name, setName] = useState(user?.name ?? "")
  const [username, setUsername] = useState(user?.username ?? "")

  // Password change
  const [oldPwd, setOldPwd] = useState("")
  const [newPwd, setNewPwd] = useState("")
  const [confirmPwd, setConfirmPwd] = useState("")

  useEffect(() => {
    if (user) {
      setName(user.name ?? "")
      setUsername(user.username ?? "")
    }
  }, [user])

  const updateMut = useMutation({
    mutationFn: (body: UpdateProfileBody) => sendBg({ type: "api.updateProfile", body }),
    onSuccess: (data) => {
      setUser(data)
      toast({ title: "Профиль обновлён", variant: "success" })
    },
    onError: (err: Error) => {
      toast({
        title: "Не удалось сохранить",
        description: err.message,
        variant: "error",
      })
    },
  })

  const changePwdMut = useMutation({
    mutationFn: (vars: { oldPassword: string; newPassword: string }) =>
      sendBg({ type: "api.changePassword", ...vars }),
    onSuccess: () => {
      toast({ title: "Пароль изменён", variant: "success" })
      setOldPwd("")
      setNewPwd("")
      setConfirmPwd("")
    },
    onError: (err: Error) => {
      toast({ title: "Не удалось изменить пароль", description: err.message, variant: "error" })
    },
  })

  function handleProfileSave() {
    updateMut.mutate({ name: name.trim(), username: username.trim() })
  }

  function handlePwdSave() {
    if (newPwd.length < 8) {
      toast({ title: "Минимум 8 символов", variant: "error" })
      return
    }
    if (newPwd !== confirmPwd) {
      toast({ title: "Пароли не совпадают", variant: "error" })
      return
    }
    changePwdMut.mutate({ oldPassword: oldPwd, newPassword: newPwd })
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Профиль</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-5">
        {/* Email read-only */}
        <section className="space-y-1">
          <Label>Email</Label>
          <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 px-3 py-2 text-xs">
            {user?.email ?? "—"}
          </div>
          <p className="text-[10px] text-(--color-muted-foreground)">
            Email нельзя изменить здесь — пишите в поддержку.
          </p>
        </section>

        {/* Name + username */}
        <section className="space-y-3">
          <div className="space-y-1">
            <Label htmlFor="prof-name">Имя</Label>
            <Input
              id="prof-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
              placeholder="Ваше имя"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="prof-username">Username</Label>
            <Input
              id="prof-username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={30}
              placeholder="username"
            />
            <p className="text-[10px] text-(--color-muted-foreground)">
              Только латиница, цифры, _. До 30 символов.
            </p>
          </div>
          <Button
            type="button"
            variant="brand"
            onClick={handleProfileSave}
            disabled={updateMut.isPending}
            className="w-full gap-1.5"
          >
            <Save className="h-3.5 w-3.5" />
            Сохранить
          </Button>
        </section>

        {/* Password change */}
        <section className="space-y-3">
          <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
            Смена пароля
          </div>
          <div className="space-y-1">
            <Label htmlFor="old-pwd">Текущий пароль</Label>
            <Input
              id="old-pwd"
              type="password"
              value={oldPwd}
              onChange={(e) => setOldPwd(e.target.value)}
              placeholder="Введите текущий"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="new-pwd">Новый пароль</Label>
            <Input
              id="new-pwd"
              type="password"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              placeholder="Минимум 8 символов"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="confirm-pwd">Подтвердить новый</Label>
            <Input
              id="confirm-pwd"
              type="password"
              value={confirmPwd}
              onChange={(e) => setConfirmPwd(e.target.value)}
              placeholder="Введите ещё раз"
            />
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={handlePwdSave}
            disabled={changePwdMut.isPending || !oldPwd || !newPwd}
            className="w-full gap-1.5"
          >
            <Lock className="h-3.5 w-3.5" />
            Сменить пароль
          </Button>
        </section>
      </div>
    </div>
  )
}
