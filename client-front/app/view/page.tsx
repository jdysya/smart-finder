'use client';

import { useSearchParams } from 'next/navigation';
import { useEffect, useState, Suspense } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { useTheme } from 'next-themes';
import { materialDark, materialLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import matter from 'gray-matter';

type FileType = 'image' | 'video' | 'pdf' | 'markdown' | 'code' | 'unknown';

interface HashParams {
  page?: number;
  time?: string;
  startLine?: number;
  endLine?: number;
}

// A simple function to extract the language from a filename
const getLanguageFromFilename = (filename: string): string => {
  const extension = filename.split('.').pop();
  // You can add more mappings here if needed
  switch (extension) {
    case 'js':
      return 'javascript';
    case 'ts':
      return 'typescript';
    case 'py':
      return 'python';
    case 'go':
      return 'go';
    case 'java':
      return 'java';
    case 'html':
      return 'html';
    case 'css':
      return 'css';
    default:
      return extension || 'text';
  }
};

// Parses the URL hash for view parameters
const parseHash = (hash: string): HashParams => {
  const params: HashParams = {};
  if (!hash) return params;

  const hashContent = hash.substring(1);

  // For video time, e.g., #t=1m30s
  const timeMatch = hashContent.match(/^t=([\w\.]+)$/);
  if (timeMatch) {
    params.time = timeMatch[1];
    return params;
  }

  // For PDF page, e.g., #page=5
  const pageMatch = hashContent.match(/^page=(\d+)$/);
  if (pageMatch) {
    params.page = parseInt(pageMatch[1], 10);
    return params;
  }

  // For code lines, e.g., #L10 or #L10-L20
  const lineMatch = hashContent.match(/^L(\d+)(?:-L(\d+))?$/);
  if (lineMatch) {
    params.startLine = parseInt(lineMatch[1], 10);
    params.endLine = lineMatch[2] ? parseInt(lineMatch[2], 10) : params.startLine;
  }

  return params;
};

// Component to display frontmatter data
const FrontmatterDisplay = ({ data }: { data: { [key: string]: any } }) => {
  if (Object.keys(data).length === 0) {
    return null;
  }

  return (
    <div className="mb-6 p-4 border rounded-lg bg-gray-50 dark:bg-gray-800 not-prose">
      <h2 className="text-xl font-bold mb-2">Metadata</h2>
      <dl className="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-2">
        {Object.entries(data).map(([key, value]) => (
          <div key={key} className="flex flex-col">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">{key}</dt>
            <dd className="mt-1 text-base">
              {Array.isArray(value)
                ? value.map((item, index) => (
                    <span
                      key={index}
                      className="inline-block bg-gray-200 dark:bg-gray-700 rounded-full px-3 py-1 text-sm font-semibold mr-2 mb-2"
                    >
                      {String(item)}
                    </span>
                  ))
                : String(value)}
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
};

// Component that uses useSearchParams
function ViewPageContent() {
  const searchParams = useSearchParams();
  const { theme } = useTheme();
  const hash = searchParams.get('hash');

  const [fileUrl, setFileUrl] = useState('');
  const [rawContent, setRawContent] = useState('');
  const [frontmatter, setFrontmatter] = useState<{ [key: string]: any }>({});
  const [markdownContent, setMarkdownContent] = useState('');
  const [fileType, setFileType] = useState<FileType>('unknown');
  const [filename, setFilename] = useState('');
  const [hashParams, setHashParams] = useState<HashParams>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setHashParams(parseHash(window.location.hash));
    }

    if (!hash) {
      setError('No hash provided in the URL.');
      setLoading(false);
      return;
    }

    const url = `/api/md5?hash=${hash}`;
    setFileUrl(url);

    const fetchData = async () => {
      try {
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const disposition = response.headers.get('Content-Disposition');
        if (disposition && disposition.includes('filename=')) {
          const match = /filename="([^"]+)"/.exec(disposition);
          if (match && match[1]) {
            setFilename(match[1]);
          }
        }

        const contentType = response.headers.get('Content-Type') || '';

        if (contentType.startsWith('image/')) {
          setFileType('image');
        } else if (contentType.startsWith('video/')) {
          setFileType('video');
        } else if (contentType === 'application/pdf') {
          setFileType('pdf');
        } else if (contentType.includes('text/markdown')) {
          setFileType('markdown');
          const text = await response.text();
          const { data, content } = matter(text);
          setFrontmatter(data);
          setMarkdownContent(content);
        } else if (contentType.startsWith('text/')) {
          setFileType('code');
          const text = await response.text();
          setRawContent(text);
        } else {
          setFileType('unknown');
          try {
            const text = await response.text();
            setRawContent(text);
          } catch (e) {
            setRawContent('Cannot display binary content.');
          }
        }
      } catch (e: any) {
        setError(e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [hash]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  const renderContent = () => {
    const { startLine, endLine } = hashParams;
    const shouldHighlight = !!(fileType === 'code' && startLine && endLine);

    switch (fileType) {
      case 'image':
        return <img src={fileUrl} alt={filename} className="max-w-full h-auto mx-auto" />;
      case 'video':
        return <video src={`${fileUrl}${window.location.hash}`} controls className="max-w-full mx-auto" />;
      case 'pdf':
        return (
          <object data={`${fileUrl}${window.location.hash}`} type="application/pdf" width="100%" height="1000px">
            <p>
              This browser does not support PDFs. Please download the PDF to view it:{' '}
              <a href={fileUrl}>Download PDF</a>
            </p>
          </object>
        );
      case 'markdown':
        return (
          <>
            <FrontmatterDisplay data={frontmatter} />
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdownContent}</ReactMarkdown>
          </>
        );
      case 'code':
        return (
          <SyntaxHighlighter
            language={getLanguageFromFilename(filename)}
            style={theme === 'dark' ? materialDark : materialLight}
            showLineNumbers
            wrapLines={shouldHighlight}
            lineProps={lineNumber => {
              if (shouldHighlight && lineNumber >= startLine && lineNumber <= endLine) {
                return { style: { display: 'block', backgroundColor: 'rgba(255, 255, 255, 0.1)' } };
              }
              return {};
            }}
          >
            {rawContent}
          </SyntaxHighlighter>
        );
      default:
        return <pre>{rawContent}</pre>;
    }
  };

  return (
    <div className="container mx-auto p-4">
      {fileType === 'markdown' ? (
        <article className="prose dark:prose-invert max-w-none">{renderContent()}</article>
      ) : (
        renderContent()
      )}
    </div>
  );
}

// Loading component for Suspense fallback
function ViewPageLoading() {
  return <div className="container mx-auto p-4">Loading...</div>;
}

export default function ViewPage() {
  return (
    <Suspense fallback={<ViewPageLoading />}>
      <ViewPageContent />
    </Suspense>
  );
}

