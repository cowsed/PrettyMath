

#version 330
in vec2 UV;
uniform float u_time;
uniform float time;
uniform float paramB;
uniform float numTiles;
uniform vec3 offset;
uniform vec3 size;
uniform float paramA;

uniform vec3 col1;//Light green
uniform vec3 col2;//Dark Green

uniform float mix_before_after;

layout(location = 0) out vec4 frag_colour;

mat2 rotate2d(float _angle){
    return mat2(cos(_angle),-sin(_angle),
                sin(_angle),cos(_angle));
}

float box(in vec2 _st, in vec2 _size){
    _size = vec2(0.5) - _size*0.5;
    vec2 uv = smoothstep(_size,
                        _size+vec2(0.001),
                        _st);
    uv *= smoothstep(_size,
                    _size+vec2(0.001),
                    vec2(1.0)-_st);
    return uv.x*uv.y;
}
float circle(in vec2 _st, in float r){
    return length(_st)-r;
}

void main() {
    vec2 uv = UV;
    vec2 uv2 = UV;
    uv.x/=(8.0/6.0);
    uv.y-=time*paramB;
    uv=fract(uv*numTiles/2)-.5;


    float ang=time*paramA;    
    uv*=rotate2d(ang);
    
    uv+=offset.xy;
    //uv.x+=.25*sin(ang);
    vec3 colf1 = vec3(uv,0);
    
    
    
    
    colf1.xy=abs(uv);
    float f=step(box(uv,size.xy),0);
    colf1=mix(col1,col2,f);
    
    vec3 colf2 = vec3(0,0,0);
    
    uv2.x/=(8.0/6.0);
    
    uv2*=numTiles/2;
    float t=time/10.0;
    vec2 uv3=uv2;
    uv2+=vec2(cos(t),sin(t));
    uv3+=vec2(-cos(-t),-sin(-t));

    uv2=fract(uv2)-.5;    
    uv3=fract(uv3)-.5;    
    
    colf2=vec3(clamp(step(circle(uv2,.1),.01)+step(circle(uv3,.1),.01),0,1));
    colf1=mix(colf1,vec3(1,0,0),colf2.x);
    vec3 col = mix(colf1,colf2, mix_before_after);
    frag_colour = vec4(col, 1);
}
