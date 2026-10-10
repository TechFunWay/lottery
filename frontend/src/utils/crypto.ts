import { md5 } from 'js-md5'

/**
 * 前端密码加密：对原始密码做一次 MD5
 * 传输到后端后，后端再做一次带盐 MD5(前端md5结果 + 'caipiao')
 */
export function hashPassword(password: string): string {
  return md5(password)
}
