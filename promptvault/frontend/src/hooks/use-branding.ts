import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { fetchBranding, updateBranding, type BrandingInput } from "@/api/branding"

export function useBranding(slug: string, enabled = true) {
  return useQuery({
    queryKey: ["branding", slug],
    queryFn: () => fetchBranding(slug),
    enabled: enabled && slug.length > 0,
  })
}

export function useUpdateBranding(slug: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: BrandingInput) => updateBranding(slug, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branding", slug] })
    },
  })
}
