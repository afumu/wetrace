import React from 'react';
import { ShieldCheck, ShieldAlert, Lock, Database, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface AgreementModalProps {
  onAccept: () => void;
}

export const AgreementModal: React.FC<AgreementModalProps> = ({ onAccept }) => {
  return (
    <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-background/80 backdrop-blur-md p-4 animate-in fade-in duration-500">
      <div className="bg-card border shadow-2xl rounded-2xl w-full max-w-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300 max-h-[90vh]">
        {/* Header */}
        <div className="p-8 border-b bg-primary/5">
          <div className="flex items-center gap-4 mb-2">
            <div className="bg-primary/10 p-3 rounded-xl text-primary">
              <ShieldCheck className="w-8 h-8" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight">使用协议与隐私声明</h2>
              <p className="text-muted-foreground text-sm mt-1">请在开始使用前仔细阅读以下条款</p>
            </div>
          </div>
        </div>

        {/* Content */}
        <div 
          className="p-8 overflow-y-auto space-y-8 text-sm leading-relaxed"
        >
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-primary font-bold">
              <ShieldAlert className="w-4 h-4" />
              <h4>1. 隐私安全警示</h4>
            </div>
            <p className="text-muted-foreground pl-6">
              本工具涉及的数据库文件、密钥信息及解析出的聊天记录均属于<span className="text-foreground font-bold underline decoration-primary/30 underline-offset-4">极度敏感的个人隐私数据</span>。请务必妥善保管相关文件。
            </p>
          </section>

          <section className="space-y-3">
            <div className="flex items-center gap-2 text-primary font-bold">
              <Lock className="w-4 h-4" />
              <h4>2. 严禁传播承诺</h4>
            </div>
            <div className="text-muted-foreground pl-6 space-y-2">
              <p>
                <span className="text-foreground font-bold italic">您承诺：</span> 仅将本工具用于个人数据备份、学习研究或合法取证用途。
              </p>
              <p className="bg-destructive/5 border-l-2 border-destructive p-3 rounded-r-lg">
                <span className="text-destructive font-bold">严禁</span>将获取到的任何敏感文件、密钥或聊天记录向外传播、上传至任何第三方平台或用于任何非法用途。由此产生的任何隐私泄露或法律后果，均由用户本人承担。
              </p>
            </div>
          </section>

          <section className="space-y-3">
            <div className="flex items-center gap-2 text-primary font-bold">
              <Database className="w-4 h-4" />
              <h4>3. 数据本地化保障</h4>
            </div>
            <p className="text-muted-foreground pl-6 italic">
              <span className="text-foreground font-bold not-italic">本工具的所有数据处理（包括解密、分析、存储等）均完全在您的本地计算机上完成。</span> 软件不会向任何外部服务器、云端或第三方机构上传或共享您的任何数据，请放心使用。
            </p>
          </section>

          <section className="space-y-3">
            <div className="flex items-center gap-2 text-primary font-bold">
              <ShieldCheck className="w-4 h-4" />
              <h4>4. 责任限制</h4>
            </div>
            <p className="text-muted-foreground pl-6">
              本工具仅作为技术辅助工具提供，开发者不对因使用本程序导致的任何数据风险、系统稳定性问题或法律纠纷承担责任。
            </p>
          </section>

          <div className="pt-4 text-center text-xs text-muted-foreground border-t border-dashed">
             继续使用即表示您已完全理解并同意上述所有条款
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-muted/20 flex justify-center">
          <Button 
            size="lg" 
            className="rounded-full px-12 py-6 text-lg font-bold shadow-lg hover:shadow-primary/20 transition-all gap-2"
            onClick={onAccept}
          >
            我已阅读并同意上述协议
            <ArrowRight className="w-5 h-5" />
          </Button>
        </div>
      </div>
    </div>
  );
};
