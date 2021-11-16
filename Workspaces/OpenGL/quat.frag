#version 330
in vec2 UV;

layout(location = 0) out vec4 frag_colour;

uniform int MAX_ITER; 

uniform float sliceX;
uniform float sliceY;
uniform float sliceZ;
uniform float sliceW;


uniform float scaleAll;
uniform float scaleX;
uniform float scaleY;

uniform float radius;

vec4 mulQuat(vec4 q1, vec4 q2){
    vec4 q;
    q.x = (q1.w * q2.x) + (q1.x * q2.w) + (q1.y * q2.z) - (q1.z * q2.y);
    q.y = (q1.w * q2.y) - (q1.x * q2.z) + (q1.y * q2.w) + (q1.z * q2.x);
    q.z = (q1.w * q2.z) + (q1.x * q2.y) - (q1.y * q2.x) + (q1.z * q2.w);
    q.w = (q1.w * q2.w) - (q1.x * q2.x) - (q1.y * q2.y) - (q1.z * q2.z);

    return q;
}

vec2 mulCmplx(vec2 a, vec2 b){
    vec2 c;
    c.x = a.x*b.x - a.y*b.y;
    c.y = abs(-2 * a.y * b.x);
    return c;
}



int iterate(vec2 uv){
    int i;
    vec4 c = vec4(uv.x , sliceW, sliceZ, uv.y);
    vec4 z = c;
    //vec2 c = uv;
    //vec2 z = uv;
    
    for (i=0; i<MAX_ITER; i++){
         z = mulQuat(z, z) + c;
         if (length(z)>radius){
             break;
         }
             
    }
    return i;
}

void main() {
    float aspect = 8.0/6.0;
    vec2 uv = UV/vec2(aspect, 1);
    uv+=vec2(sliceX, sliceY);
    uv*=vec2(scaleX, scaleY);
    uv/=scaleAll;
    uv=vec2(uv.y,uv.x);
    
    
    int iters = iterate(uv);
    
    float amt = float(iters)/float(MAX_ITER);
    
    vec3 col = vec3(amt);
    //col*=texture(tex, uv.xy).rgb;
    frag_colour = vec4(col, 1);
}
