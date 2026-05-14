import { useCallback, useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  getWorkspace,
  onWorkspaceChanged,
  setCollection,
  setWorkspace,
  type WorkspaceSelection,
} from '../lib/storage';
import { sendBg } from '../lib/bg-client';
import type { CollectionDTO, TeamDTO } from '../lib/types';

export interface WorkspaceState extends WorkspaceSelection {
  teams: TeamDTO[];
  collections: CollectionDTO[];
  // Текущая команда (lookup по workspaceId в teams[]).
  //   null      — workspaceId=null (явный personal workspace)
  //   undefined — workspaceId есть, но teamsQuery ещё грузится (UI должен
  //               показывать skeleton, а не «Личное»)
  //   TeamDTO   — найдена команда
  //   null      — команда не найдена (юзера выгнали из команды)
  currentTeam: TeamDTO | null | undefined;
  isLoading: boolean;
  setWorkspaceId: (id: number | null) => void;
  setCollectionId: (id: number | null) => void;
}

export function useWorkspace(): WorkspaceState {
  const [selection, setSelection] = useState<WorkspaceSelection>({
    workspaceId: null,
    collectionId: null,
  });
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void getWorkspace().then((s) => {
      if (!cancelled) {
        setSelection(s);
        setHydrated(true);
      }
    });
    const off = onWorkspaceChanged((s) => {
      if (!cancelled) setSelection(s);
    });
    return () => {
      cancelled = true;
      off();
    };
  }, []);

  const teamsQuery = useQuery({
    queryKey: ['workspace', 'teams'],
    queryFn: () => sendBg({ type: 'api.listTeams' }),
    enabled: hydrated,
    staleTime: 5 * 60 * 1000,
  });

  const collectionsQuery = useQuery({
    queryKey: ['workspace', 'collections', selection.workspaceId],
    queryFn: () =>
      sendBg({ type: 'api.listCollections', teamId: selection.workspaceId ?? null }),
    enabled: hydrated,
    staleTime: 2 * 60 * 1000,
  });

  const setWorkspaceId = useCallback((id: number | null) => {
    void setWorkspace(id);
  }, []);

  const setCollectionId = useCallback((id: number | null) => {
    void setCollection(id);
  }, []);

  const teams = teamsQuery.data ?? [];
  // currentTeam: пока teams грузится — undefined (unknown), чтобы UI не
  // мигал «Личное». null означает явный personal workspace или потерянную
  // команду (kicked out), и UI обрабатывает это иначе.
  let currentTeam: TeamDTO | null | undefined;
  if (selection.workspaceId == null) {
    currentTeam = null;
  } else if (teamsQuery.isPending && !teamsQuery.data) {
    currentTeam = undefined;
  } else {
    currentTeam = teams.find((t) => t.id === selection.workspaceId) ?? null;
  }

  return {
    workspaceId: selection.workspaceId,
    collectionId: selection.collectionId,
    teams,
    collections: collectionsQuery.data ?? [],
    currentTeam,
    isLoading: teamsQuery.isPending || collectionsQuery.isPending,
    setWorkspaceId,
    setCollectionId,
  };
}
