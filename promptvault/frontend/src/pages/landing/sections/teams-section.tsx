import { Users, Mail, FolderOpen, Shield } from "lucide-react"
import { FadeIn } from "../components/fade-in"
import { teamsContent } from "../data/landing-content"

export function TeamsSection() {
  return (
    <section className="py-16 sm:py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="grid items-center gap-12 lg:grid-cols-2">
          {/* Left: text */}
          <FadeIn direction="left">
            <div>
              <div className="mb-4 inline-flex items-center gap-2 rounded-full bg-violet-500/10 px-3 py-1 text-sm text-violet-300">
                <Users className="h-3.5 w-3.5" />
                Совместная работа
              </div>
              <h2 className="text-3xl font-bold sm:text-4xl">{teamsContent.title}</h2>
              <p className="mt-4 text-muted-foreground leading-relaxed">{teamsContent.desc}</p>

              <ul className="mt-6 space-y-2">
                {teamsContent.bullets.map(b => (
                  <li key={b} className="flex items-center gap-2 text-sm text-muted-foreground">
                    <div className="h-1 w-1 rounded-full bg-violet-400/60" />
                    {b}
                  </li>
                ))}
              </ul>
            </div>
          </FadeIn>

          {/* Right: roles illustration */}
          <FadeIn direction="right" delay={150}>
            <div className="rounded-2xl border border-border/30 bg-card/20 p-6" aria-hidden="true">
              {/* Roles visualization */}
              <div className="space-y-3">
                {teamsContent.roles.map((role, i) => (
                  <div
                    key={role.name}
                    className="flex items-center gap-4 rounded-xl border border-border/20 bg-background/30 px-4 py-3 transition-all duration-300 hover:border-violet-500/10"
                    style={{
                      animationDelay: `${i * 200}ms`,
                    }}
                  >
                    {/* Avatar */}
                    <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-violet-500/20 to-violet-600/5 ring-1 ring-violet-500/10">
                      <span className="text-xs font-medium text-violet-300">
                        {role.name[0]}
                      </span>
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">{role.name}</span>
                        <span className={`rounded-md px-1.5 py-0.5 text-[0.55rem] font-medium ${role.color} bg-current/10`}>
                          {role.desc}
                        </span>
                      </div>
                    </div>
                    {i === 0 && <Shield className="h-3.5 w-3.5 text-violet-400/40" />}
                  </div>
                ))}
              </div>

              {/* Invite animation hint */}
              <div className="mt-4 flex items-center gap-2 rounded-lg border border-dashed border-border/30 bg-background/20 px-3 py-2.5">
                <Mail className="h-3.5 w-3.5 text-muted-foreground/40" />
                <span className="text-xs text-muted-foreground/40">Пригласить по почте...</span>
              </div>

              {/* Team collection */}
              <div className="mt-3 flex items-center gap-2 rounded-lg border border-border/20 bg-background/20 px-3 py-2.5">
                <FolderOpen className="h-3.5 w-3.5 text-blue-400/40" />
                <span className="text-xs text-muted-foreground/50">Командная коллекция</span>
                <span className="ml-auto text-[0.55rem] text-muted-foreground/30">3 промпта</span>
              </div>
            </div>
          </FadeIn>
        </div>
      </div>
    </section>
  )
}
