## sector-sanity-checker-v0.3.0

这个工具可以帮助检测window post阶段错误的filecash sector.
核心逻辑:模拟生成windowpost,然后在验证windowpost。若果有异常，会弹出终端执行，如果没有异常，返回nil。

func (sb *Sealer) GenerateWindowPoSt(ctx context.Context, minerID abi.ActorID, sectorInfo []proof2.SectorInfo, randomness abi.PoStRandomness) 


func (proofVerifier) VerifyWindowPoSt(ctx context.Context, info proof2.WindowPoStVerifyInfo) (bool, error)

## 下载

http://git.pocyc.com/rectinajh/filecash_check.git

## 使用
### 步骤一, 设置环境变量

- export FIL_PROOFS_PARENT_CACHE=<YOUR_PARENT_CACHE>
- export FIL_PROOFS_PARAMETER_CACHE=<YOUR_FIL_PROOFS_PARAMETER_CACHE>
- export FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1 
- export RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1 
- export FIL_PROOFS_MAXIMIZE_CACHING=1
- export MINER_API_INFO=<YOUR_MINER_API_INFO>

 例如：

- export FIL_PROOFS_PARENT_CACHE=/mnt/proofs_parent_cache
- export FIL_PROOFS_PARAMETER_CACHE=/mnt/proofs/
- export FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1
- export RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1
- export FIL_PROOFS_MAXIMIZE_CACHING=1
- export MINER_API_INFO=/ip4/127.0.0.1/tcp/2345/http

### 步骤二, 运行filecash-check 

   注意：需要将封装文件的目录下才可扫描到扇区。
   
   在sealed下面，运行下面的脚本，获取扫描的sectorID和commonr,文件注意不要保存到sealed目录下面,其他home目录即可。文件名默认是sectors.txt。
   
 -   ll | awk '{print $9}' | awk -F '-' '{print $3}' |xargs -i lotus-miner sectors status {} | grep CIDcommR -B3 | grep -v "Status\|CIDcommD\|--" | awk '{print $2}' | tee xxx.txt

命令：

 - $>filecash-check checking  --sector-size=4G --sectors-file=/home/test/sectors.txt --miner-addr=t### --storage-dir=/opt/data/storage
 - --sector-size 是默认扫描proof证明文件
 - --sectors-file是默认扫描文件名为sectors.txt的文件,可以根据文件路径添加.
案例：


 
 
 所有的扇区在/opt/data/storage/sealed/s-xxxxx-xxx将被扫描
   
   

## filecash-checker-

This tools can help you check sector to avoid the window PoST fail.

## Download

http://git.pocyc.com/rectinajh/filecash_check.git

## Usage
### step 1, export the environment variable
 - export FIL_PROOFS_PARENT_CACHE=<YOUR_PARENT_CACHE>
 - export FIL_PROOFS_PARAMETER_CACHE=<YOUR_FIL_PROOFS_PARAMETER_CACHE>
 - export FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1 
 - export RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1 
 - export FIL_PROOFS_MAXIMIZE_CACHING=1
 - export MINER_API_INFO=<YOUR_MINER_API_INFO>

 
### step 2, run the tool 
 - filecash-check checking  --sector-size=32G --miner-addr=<your_miner_id> --storage-dir= <sector_dir> 

 
### For Example:

 - filecash-check checking  --sector-size=32G --miner-addr=t### --storage-dir=/opt/data/storage
 
 Then all the sectors under /opt/data/storage/sealed/s-xxxxx-xxx will be scaned.
 
 - filecash-check checking  --sector-size=32G --sectors-file-only-number=sectors-to-scan.txt --miner-addr=t### --storage-dir=/opt/data/storage
 
 Then all the sectors specified by sectors-to-scan.txt  under folder /opt/data/storage will be scaned. 
  