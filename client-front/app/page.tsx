'use client'
import { useEffect, useState } from 'react';
import { Button } from '@heroui/button';
import { Input } from '@heroui/input';
import { Card, CardHeader, CardBody } from '@heroui/card';
import { Progress } from '@heroui/progress';
import { Table, TableHeader, TableColumn, TableBody, TableRow, TableCell } from '@heroui/table';

export default function Home() {
  const [dirs, setDirs] = useState([]);
  const [newDir, setNewDir] = useState('');
  const [status, setStatus] = useState<any>({});
  const [files, setFiles] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [path, setPath] = useState('');
  const [conversionResult, setConversionResult] = useState('');

  const fetchDirs = () => {
    fetch('/api/directories')
      .then((res) => res.json())
      .then(setDirs);
  };

  const fetchStatus = () => {
    fetch('/api/status')
      .then((res) => res.json())
      .then(setStatus);
  };

  const fetchFiles = () => {
    fetch('/api/files')
      .then((res) => res.json())
      .then(setFiles);
  };

  useEffect(() => {
    fetchDirs();
    fetchStatus();
    fetchFiles();
    const interval = setInterval(fetchStatus, 1500);
    return () => clearInterval(interval);
  }, []);

  const addDir = () => {
    fetch('/api/directories', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: newDir }),
    }).then(() => {
      setNewDir('');
      fetchDirs();
    });
  };

  const delDir = (path) => {
    fetch('/api/directories', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    }).then(fetchDirs);
  };

  const convertPath = () => {
    fetch('/api/path2url', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    })
      .then(async (res) => {
        if (!res.ok) {
          const msg = await res.text();
          setConversionResult(`Error: ${msg}`);
        } else {
          const data = await res.json();
          setConversionResult(`MD5: ${data.md5} | URL: ${data.url}`);
        }
      })
      .catch((err) => setConversionResult(`Error: ${err.message}`));
  };

  const filteredFiles = files.filter(
    (file) =>
      file.filename.toLowerCase().includes(searchTerm.toLowerCase()) ||
      file.path.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Smart Finder</h1>

      <Card className="mb-4">
        <CardHeader>监控目录</CardHeader>
        <CardBody>
          <div className="flex gap-2 mb-2">
            <Input
              value={newDir}
              onChange={(e) => setNewDir(e.target.value)}
              placeholder="输入目录路径"
            />
            <Button onClick={addDir}>添加目录</Button>
          </div>
          <ul>
            {dirs.map((dir) => (
              <li key={dir} className="flex justify-between items-center p-2 bg-gray-100 rounded mb-1">
                <span>{dir}</span>
                <Button color="danger" size="sm" onClick={() => delDir(dir)}>
                  删除
                </Button>
              </li>
            ))}
          </ul>
        </CardBody>
      </Card>

      <Card className="mb-4">
        <CardHeader>状态</CardHeader>
        <CardBody>
          {status.indexing ? (
            <div>
              <p>
                Indexing: {status.indexingDone} / {status.indexingTotal} files
              </p>
              <Progress
                value={(status.indexingDone / status.indexingTotal) * 100}
              />
            </div>
          ) : (
            <p>索引数量: {status.fileCount}</p>
          )}
        </CardBody>
      </Card>

      <Card className="mb-4">
        <CardHeader>路径转换</CardHeader>
        <CardBody>
          <div className="flex gap-2 mb-2">
            <Input
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder="输入文件完整路径"
            />
            <Button onClick={convertPath}>转换</Button>
          </div>
          {conversionResult && <p>{conversionResult}</p>}
        </CardBody>
      </Card>

      <Card>
        <CardHeader>索引文件</CardHeader>
        <CardBody>
          <Input
            className="mb-2"
            placeholder="搜索文件..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          <Table aria-label="Indexed Files">
            <TableHeader>
              <TableColumn>文件名</TableColumn>
              <TableColumn>路径</TableColumn>
              <TableColumn>大小</TableColumn>
              <TableColumn>修改时间</TableColumn>
              <TableColumn>操作</TableColumn>
            </TableHeader>
            <TableBody>
              {filteredFiles.map((file) => (
                <TableRow key={file.md5}>
                  <TableCell>{file.filename}</TableCell>
                  <TableCell>{file.path}</TableCell>
                  <TableCell>{(file.size / 1024).toFixed(2)} KB</TableCell>
                  <TableCell>{new Date(file.modified_at).toLocaleString()}</TableCell>
                  <TableCell>
                    <Button as="a" href={`/direct/md5?hash=${file.md5}`} target="_blank" size="sm">
                      打开
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardBody>
      </Card>
    </div>
  );
}