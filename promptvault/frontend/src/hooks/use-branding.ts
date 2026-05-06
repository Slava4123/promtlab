import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  deleteLogo,
  fetchBranding,
  updateBranding,
  uploadLogo,
  type BrandingInput,
} from "@/api/branding"

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

export function useUploadLogo(slug: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => uploadLogo(slug, file),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branding", slug] })
    },
  })
}

export function useDeleteLogo(slug: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => deleteLogo(slug),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branding", slug] })
    },
  })
}
