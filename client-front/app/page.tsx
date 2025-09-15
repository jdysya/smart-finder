'use client'
import { useEffect, useState } from 'react';
import { Button } from '@heroui/button';
import { Input, Textarea } from '@heroui/input';
import { Card, CardHeader, CardBody } from '@heroui/card';
import { Progress } from '@heroui/progress';
import { Table, TableHeader, TableColumn, TableBody, TableRow, TableCell } from '@heroui/table';

export default function Home() {
  type FileItem = {
    md5: string;
    filename: string;
    path: string;
    size: number;
    modified_at: string;
  };

  const [dirs, setDirs] = useState([]);
  const [newDir, setNewDir] = useState('');
  const [status, setStatus] = useState<any>({});
  const [files, setFiles] = useState<FileItem[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [ignorePatterns, setIgnorePatterns] = useState('');
  const [scanStatus, setScanStatus] = useState<any>({});

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
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    });
    if (searchTerm) params.append('search', searchTerm);
    fetch(`/api/files?${params.toString()}`)
      .then((res) => res.json())
      .then((data) => {
        setFiles(data.files || []);
        setTotal(data.total || 0);
      });
  };

  const fetchIgnorePatterns = () => {
    fetch('/api/ignore-patterns')
      .then((res) => res.json())
      .then((data) => setIgnorePatterns(data?.join('\n')));
  };

  const fetchScanStatus = () => {
    fetch('/api/scan/status')
      .then((res) => res.json())
      .then(setScanStatus)
      .catch(() => setScanStatus({}));
  };

  const triggerScan = () => {
    fetch('/api/scan/trigger', {
      method: 'POST',
    }).then(() => {
      alert('扫描已触发');
      fetchScanStatus();
    }).catch(() => {
      alert('触发扫描失败');
    });
  };

  useEffect(() => {
    fetchDirs();
    fetchStatus();
    fetchFiles();
    fetchIgnorePatterns();
    fetchScanStatus();
    const interval = setInterval(fetchStatus, 1500);
    const scanInterval = setInterval(fetchScanStatus, 2000);
    return () => {
      clearInterval(interval);
      clearInterval(scanInterval);
    };
  }, []);

  useEffect(() => {
    fetchFiles();
  }, [page, pageSize, searchTerm]);

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

  const delDir = (path: string) => {
    fetch('/api/directories', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    }).then(fetchDirs);
  };

  const saveIgnorePatterns = () => {
    fetch('/api/ignore-patterns', {
      method: 'POST',
      headers: { 'Content-Type': 'text/plain' },
      body: ignorePatterns,
    }).then(() => alert('忽略规则已保存'));
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
              <li key={dir} className="flex justify-between items-center p-2  rounded mb-1">
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
        <CardHeader>扫描控制</CardHeader>
        <CardBody>
          <div className="flex gap-2 mb-4">
            <Button 
              onClick={triggerScan} 
              color="primary"
              disabled={scanStatus.is_scanning}
            >
              {scanStatus.is_scanning ? '扫描中...' : '手动触发扫描'}
            </Button>
          </div>
          
          {scanStatus.is_scanning && (
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>当前目录: {scanStatus.current_dir || '准备中...'}</span>
                <span>进度: {scanStatus.progress?.toFixed(1) || 0}%</span>
              </div>
              <Progress 
                value={scanStatus.progress || 0} 
                className="w-full"
              />
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>总文件数: {scanStatus.total_files || 0}</div>
                <div>已处理: {scanStatus.processed_files || 0}</div>
                <div>已跳过: {scanStatus.skipped_files || 0}</div>
                <div>错误: {scanStatus.error_files || 0}</div>
              </div>
              {scanStatus.elapsed_time && (
                <div className="text-sm text-gray-600">
                  已用时: {scanStatus.elapsed_time}
                </div>
              )}
            </div>
          )}
          
          {!scanStatus.is_scanning && scanStatus.total_files > 0 && (
            <div className="mt-4 p-3 rounded">
              <h4 className="font-medium mb-2">上次扫描结果:</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>总文件数: {scanStatus.total_files}</div>
                <div>已处理: {scanStatus.processed_files}</div>
                <div>已跳过: {scanStatus.skipped_files}</div>
                <div>删除过时: {scanStatus.deleted_files}</div>
                <div>错误: {scanStatus.error_files}</div>
                {scanStatus.start_time && (
                  <div className="col-span-2 mt-2 pt-2 border-t">
                    扫描时间: {new Date(scanStatus.start_time).toLocaleString()}
                  </div>
                )}
              </div>
            </div>
          )}
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
        <CardHeader>忽略规则</CardHeader>
        <CardBody>
          <Textarea
            value={ignorePatterns}
            onChange={(e) => setIgnorePatterns(e.target.value)}
            placeholder="输入忽略规则，一行一个"
            rows={10}
          />
          <Button onClick={saveIgnorePatterns} className="mt-2">保存规则</Button>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>索引文件</CardHeader>
        <CardBody>
          <Input
            className="mb-2"
            placeholder="搜索文件..."
            value={searchTerm}
            onChange={(e) => { setSearchTerm(e.target.value); setPage(1); }}
          />
          {/* 分页控件 */}
          <div className="flex gap-2 my-2 items-center">
            <Button size="sm" disabled={page === 1} onClick={() => setPage(page - 1)}>上一页</Button>
            <span>第 {page} 页 / 共 {Math.max(1, Math.ceil(total / pageSize))} 页</span>
            <Button size="sm" disabled={page * pageSize >= total} onClick={() => setPage(page + 1)}>下一页</Button>
            <span className="ml-4">每页</span>
            <select value={pageSize} onChange={e => { setPageSize(Number(e.target.value)); setPage(1); }} className="border rounded px-1 py-0.5">
              {[10, 20, 50, 100].map(sz => <option key={sz} value={sz}>{sz}</option>)}
            </select>
            <span>条</span>
          </div>
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
                    <Button as="a" href={`/view?hash=${file.md5}`} target="_blank" size="sm">
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