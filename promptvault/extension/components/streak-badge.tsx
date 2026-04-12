import { Flame } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { sendBg } from '../lib/bg-client';
import { Badge } from './ui/badge';

export function StreakBadge() {
  const { data } = useQuery({
    queryKey: ['streak'],
    queryFn: () => sendBg({ type: 'api.getStreak' }),
    staleTime: 60 * 1000,
    retry: 0,
  });

  if (!data || data.current_streak === 0) return null;

  return (
    <Badge
      variant="outline"
      title={`Серия: ${data.current_streak} дней (рекорд: ${data.longest_streak})`}
      className="gap-1 border-orange-500/30 bg-orange-500/10 text-orange-400"
    >
      <Flame className="h-2.5 w-2.5 fill-current" />
      {data.current_streak}
    </Badge>
  );
}
