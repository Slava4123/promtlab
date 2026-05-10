"""Стресс-тест Canvas: генерирует цепочку с глубокой вложенностью fork'ов
и большим количеством шагов через реальный API.

Запуск (из promptvault/, при поднятом docker compose):
    python scripts/stress-chain.py

Юзер: max@test.local / Test12345!  — должен существовать (см. seed в README dev).
"""
import json
import sys
import urllib.request
import urllib.error

API = "http://localhost:8080/api"


def login(email: str, password: str) -> str:
    req = urllib.request.Request(
        f"{API}/auth/login",
        data=json.dumps({"email": email, "password": password}).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req) as r:
        return json.loads(r.read())["tokens"]["access_token"]


class Chain:
    def __init__(self, token: str, name: str):
        self.token = token
        req = urllib.request.Request(
            f"{API}/chains",
            data=json.dumps({"name": name, "description": "Stress-test"}).encode(),
            headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"},
            method="POST",
        )
        with urllib.request.urlopen(req) as r:
            self.id = json.loads(r.read())["id"]

    def add(self, body: dict) -> int:
        req = urllib.request.Request(
            f"{API}/chains/{self.id}/steps",
            data=json.dumps(body).encode(),
            headers={"Content-Type": "application/json", "Authorization": f"Bearer {self.token}"},
            method="POST",
        )
        try:
            with urllib.request.urlopen(req) as r:
                return json.loads(r.read())["id"]
        except urllib.error.HTTPError as e:
            print(f"  FAILED body={body}: {e.read().decode()}", file=sys.stderr)
            raise

    def add_prompt(self, prompt_id, name, after=None, parent_fork=None, branch_idx=None):
        body = {"prompt_id": prompt_id, "name": name}
        if after is not None:
            body["after_step_id"] = after
        if parent_fork is not None:
            body["parent_fork_id"] = parent_fork
            body["branch_index"] = branch_idx
        return self.add(body)

    def add_fork(self, name, branch_labels, after=None, parent_fork=None, branch_idx=None):
        body = {
            "step_type": "fork",
            "name": name,
            "conditions": {"branches": [{"label": lb} for lb in branch_labels]},
        }
        if after is not None:
            body["after_step_id"] = after
        if parent_fork is not None:
            body["parent_fork_id"] = parent_fork
            body["branch_index"] = branch_idx
        return self.add(body)


def main():
    token = login("max@test.local", "Test12345!")
    print(f"token: {token[:24]}...")

    PROMPTS = [44, 43, 42, 41, 40, 39, 38, 37]
    pi = iter(PROMPTS * 10)

    c = Chain(token, "Stress test — deep tree")
    print(f"chain id: {c.id}")

    s1 = c.add_prompt(next(pi), "Подготовка")
    s2 = c.add_prompt(next(pi), "Анализ контекста", after=s1)

    # Fork L1: 4 ветки.
    f1 = c.add_fork("Тип задачи", ["Bug", "Feature", "Refactor", "Docs"], after=s2)
    print(f"fork L1: {f1}")

    # Bug: 2 шага -> Fork L2A (Critical / Minor) -> Minor имеет Fork L3 (High / Low)
    b1 = c.add_prompt(next(pi), "Воспроизвести", parent_fork=f1, branch_idx=0)
    b2 = c.add_prompt(next(pi), "Локализовать", after=b1)
    f2a = c.add_fork("Серьёзность", ["Critical", "Minor"], after=b2)
    cr1 = c.add_prompt(next(pi), "Hotfix", parent_fork=f2a, branch_idx=0)
    c.add_prompt(next(pi), "Уведомить on-call", after=cr1)
    m1 = c.add_prompt(next(pi), "В backlog", parent_fork=f2a, branch_idx=1)
    f3 = c.add_fork("Приоритет", ["High", "Low"], after=m1)
    c.add_prompt(next(pi), "Запланировать спринт", parent_fork=f3, branch_idx=0)
    c.add_prompt(next(pi), "Отложить", parent_fork=f3, branch_idx=1)

    # Feature: 4 prompt линейно.
    p1 = c.add_prompt(next(pi), "User stories", parent_fork=f1, branch_idx=1)
    p2 = c.add_prompt(next(pi), "PRD draft", after=p1)
    p3 = c.add_prompt(next(pi), "Tech design", after=p2)
    c.add_prompt(next(pi), "Estimate", after=p3)

    # Refactor: prompt -> Fork L2B (Inline / Extract / Rewrite).
    r1 = c.add_prompt(next(pi), "Анализ legacy", parent_fork=f1, branch_idx=2)
    f2b = c.add_fork("Стратегия", ["Inline", "Extract", "Rewrite"], after=r1)
    c.add_prompt(next(pi), "Inline refactor", parent_fork=f2b, branch_idx=0)
    c.add_prompt(next(pi), "Extract module", parent_fork=f2b, branch_idx=1)
    c.add_prompt(next(pi), "Полная переписка", parent_fork=f2b, branch_idx=2)

    # Docs: 2 prompt линейно.
    d1 = c.add_prompt(next(pi), "Outline", parent_fork=f1, branch_idx=3)
    c.add_prompt(next(pi), "Полный текст", after=d1)

    # Сводка
    req = urllib.request.Request(
        f"{API}/chains/{c.id}", headers={"Authorization": f"Bearer {token}"}
    )
    with urllib.request.urlopen(req) as r:
        data = json.loads(r.read())
    steps = data.get("steps", [])
    forks = sum(1 for s in steps if s.get("step_type") == "fork")
    print(
        f"DONE: chain id={c.id}, total steps={len(steps)}, forks={forks}, prompts={len(steps) - forks}"
    )
    print(f"open: http://localhost:5173/chains/{c.id}/canvas")


if __name__ == "__main__":
    main()
