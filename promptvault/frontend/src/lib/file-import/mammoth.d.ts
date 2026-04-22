// Тип-decl для mammoth.browser.js (пакет не предоставляет types для browser build).
declare module "mammoth/mammoth.browser.js" {
  interface ConvertMessage {
    type: string
    message: string
  }

  interface ConvertResult {
    value: string
    messages: ConvertMessage[]
  }

  interface Image {
    contentType: string
    altText?: string
    readAsArrayBuffer(): Promise<ArrayBuffer>
    readAsBase64String(): Promise<string>
  }

  interface ImageAttributes {
    src: string
    alt?: string
  }

  interface ConvertOptions {
    convertImage?: (image: Image) => Promise<ImageAttributes>
    styleMap?: string[]
    includeDefaultStyleMap?: boolean
  }

  interface Input {
    arrayBuffer: ArrayBuffer
  }

  const mammoth: {
    convertToMarkdown(input: Input, options?: ConvertOptions): Promise<ConvertResult>
    convertToHtml(input: Input, options?: ConvertOptions): Promise<ConvertResult>
    images: {
      imgElement(fn: (image: Image) => Promise<ImageAttributes>): (image: Image) => Promise<ImageAttributes>
    }
  }

  export default mammoth
}
