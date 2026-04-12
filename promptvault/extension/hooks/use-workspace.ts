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

  return {
    workspaceId: selection.workspaceId,
    collectionId: selection.collectionId,
    teams: teamsQuery.data ?? [],
    collections: collectionsQuery.data ?? [],
    isLoading: teamsQuery.isPending || collectionsQuery.isPending,
    setWorkspaceId,
    setCollectionId,
  };
}
