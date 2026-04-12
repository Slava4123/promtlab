import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../lib/utils';

const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-medium',
  {
    variants: {
      variant: {
        default: 'border-(--color-border) bg-(--color-secondary) text-(--color-secondary-foreground)',
        success: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400',
        warning: 'border-amber-500/30 bg-amber-500/10 text-amber-400',
        destructive: 'border-(--color-destructive)/30 bg-(--color-destructive)/10 text-(--color-destructive)',
        outline: 'border-(--color-border) text-(--color-muted-foreground)',
      },
    },
    defaultVariants: { variant: 'default' },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}
