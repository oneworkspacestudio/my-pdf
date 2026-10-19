type Options = { document:any; request:(path:string,opt?:any)=>Promise<any>; driveFetch:(path:string,opt?:any)=>Promise<Response>; verifyDrive:()=>Promise<any> };
const params=(value:Record<string,string>)=>new URLSearchParams(value).toString();
export function copyName(name:string,index:number){const stem=name.replace(/\.pdf$/i,'').slice(0,160);return stem+(index?' ('+index+')':'')+'.pdf'}
async function sameBytes(response:Response,blob:Blob){
 if(Number(response.headers.get('Content-Length')||0)>blob.size){await response.body?.cancel();return false}
 const reader=response.body?.getReader();if(!reader)return false;const chunks:Uint8Array[]=[];let size=0;
 try{while(true){const part=await reader.read();if(part.done)break;size+=part.value.length;if(size>blob.size){await reader.cancel();return false}chunks.push(part.value)}}finally{reader.releaseLock()}
 if(size!==blob.size)return false;
 const hash=async(b:Blob)=>new Uint8Array(await crypto.subtle.digest('SHA-256',await b.arrayBuffer()));
 const [a,b]=await Promise.all([hash(blob),hash(new Blob(chunks as any))]);return a.every((v,i)=>v===b[i]);
}
// The server lease serialises tabs. Record the destination before transferring so
// a lost HTTP response can be reconciled on retry without making duplicate files.
export async function saveDocumentToDrive(o:Options){
 const profile=await o.verifyDrive();
 let claim:any;
 try{claim=await o.request('/api/documents/'+o.document.id+'/drive-save',{method:'POST',body:JSON.stringify({action:'claim'})})}catch(e:any){if(e.status===409)return;throw e}
 const receipt=async(action:string,extra:any={})=>o.request('/api/documents/'+o.document.id+'/drive-save',{method:'POST',body:JSON.stringify({action,lease:claim.lease,...extra})});
 try{
  const blob:Blob=await o.request('/api/documents/'+o.document.id);
  const source=claim.source||profile.scopes?.[0]?.name;
  if(!source)throw Error('Drive has no available source. Ask your administrator to grant access.');
  let folder=(profile.scopes.find((s:any)=>s.name===source)?.scope||'/').replace(/\/$/,'')+'/My PDF';
  if(claim.path){
   const meta=await o.driveFetch('resources?'+params({source,path:claim.path})).catch((e:any)=>{if(e.status===404)return null;throw e});
   if(meta&&await sameBytes(await o.driveFetch('resources/download?'+params({source,file:claim.path})),blob)){await receipt('saved');return{source,path:claim.path}}
   folder=claim.path.slice(0,claim.path.lastIndexOf('/'))||'/';
  }
  let directory=await o.driveFetch('resources?'+params({source,path:folder})).catch((e:any)=>{if(e.status===404)return null;throw e});
  if(!directory){
   await o.verifyDrive();
   await o.driveFetch('resources?'+params({source,path:folder,isDir:'true',override:'false'}),{method:'POST'}).catch((e:any)=>{if(e.status!==409)throw e});
   directory=await o.driveFetch('resources?'+params({source,path:folder}));
  }
  if((await directory.json()).type!=='directory')throw Error('The My PDF destination is a file. Choose another folder in Drive.');
  for(let n=0;n<1000;n++){
   const path=folder.replace(/\/$/,'')+'/'+copyName(claim.path?.split('/').pop()||o.document.name,n);
   await receipt('target',{source,path});
   await o.verifyDrive();
   try{await o.driveFetch('resources?'+params({source,path,override:'false'}),{method:'POST',headers:{'Content-Type':'application/pdf'},body:blob})}
   catch(e:any){if(e.status===409)continue;throw e}
   await receipt('saved');return{source,path};
  }
  throw Error('Too many copies with this filename. Rename files in the My PDF folder and retry.');
 }catch(e:any){await receipt(e.connectURL?'defer':'failed').catch(()=>{});throw e}
}
